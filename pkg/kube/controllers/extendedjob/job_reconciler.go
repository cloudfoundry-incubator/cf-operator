package extendedjob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ejapi "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedjob/v1alpha1"
)

type setReferenceFunc func(owner, object metav1.Object, scheme *runtime.Scheme) error

// NewJobReconciler returns a new Reconciler
func NewJobReconciler(log *zap.SugaredLogger, mgr manager.Manager) (reconcile.Reconciler, error) {
	kubeclient, err := corev1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, errors.Wrap(err, "Could not get kube client")
	}
	return &ReconcileJob{
		log:        log,
		client:     mgr.GetClient(),
		kubeclient: kubeclient,
		scheme:     mgr.GetScheme(),
	}, nil
}

// ReconcileJob reconciles an Job object
type ReconcileJob struct {
	client     client.Client
	kubeclient *corev1client.CoreV1Client
	scheme     *runtime.Scheme
	log        *zap.SugaredLogger
}

// Reconcile reads that state of the cluster for a Job object that is owned by an ExtendedJob and
// makes changes based on the state read and what is in the ExtendedJob.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Infof("Reconciling Job %s in the ExtendedJob context", request.NamespacedName)

	instance := &batchv1.Job{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.log.Info("Skip reconcile: Job not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.log.Info("Error reading the object")
		return reconcile.Result{}, err
	}

	// Get the job's extended job parent
	parentName := ""
	for _, owner := range instance.GetOwnerReferences() {
		if *owner.Controller {
			parentName = owner.Name
		}
	}
	if parentName == "" {
		r.log.Errorf("Could not find parent ExtendedJob for Job %s", request.NamespacedName)
		return reconcile.Result{}, fmt.Errorf("Could not find parent ExtendedJob for Job %s", request.NamespacedName)
	}

	ej := ejapi.ExtendedJob{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: parentName, Namespace: instance.GetNamespace()}, &ej)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Getting parent ExtendedJob")
	}

	// Persist output if needed
	if (ejapi.Output{}) != ej.Spec.Output {
		err = r.persistOutput(instance, &ej.Spec.Output)
		if err != nil {
			r.log.Errorf("Could not persist output: %s", err)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileJob) persistOutput(instance *batchv1.Job, conf *ejapi.Output) error {
	// Get job's pod. Only single-pod jobs are supported when persisting the output, so we just get the first one.
	selector, err := labels.Parse("job-name=" + instance.Name)
	if err != nil {
		return err
	}

	list := &corev1.PodList{}
	err = r.client.List(
		context.TODO(),
		&client.ListOptions{
			Namespace:     instance.GetNamespace(),
			LabelSelector: selector,
		},
		list)
	if err != nil {
		errors.Wrap(err, "Getting job's pods")
	}
	if len(list.Items) == 0 {
		errors.Errorf("Job does not own any pods?")
	}
	pod := list.Items[0]

	// Iterate over the pod's containers and store the output
	for _, c := range pod.Spec.Containers {
		options := corev1.PodLogOptions{
			Container: c.Name,
		}
		result, err := r.kubeclient.Pods(instance.GetNamespace()).GetLogs(pod.Name, &options).DoRaw()
		if err != nil {
			errors.Wrap(err, "Getting pod output")
		}

		var data map[string]string
		err = json.Unmarshal(result, &data)
		if err != nil {
			errors.Wrap(err, "Invalid output format")
		}

		// Create secret and persist the output
		secretName := conf.NamePrefix + c.Name
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: instance.GetNamespace(),
			},
			StringData: data,
		}
		err = r.client.Create(context.TODO(), secret)
	}
	return nil
}
