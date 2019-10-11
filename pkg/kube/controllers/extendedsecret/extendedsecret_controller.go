package extendedsecret

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	credsgen "code.cloudfoundry.org/cf-operator/pkg/credsgen/in_memory_generator"
	esv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedsecret/v1alpha1"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
)

// AddExtendedSecret creates a new ExtendedSecrets controller to watch for the
// custom resource and reconcile it into k8s secrets.
func AddExtendedSecret(ctx context.Context, config *config.Config, mgr manager.Manager) error {
	ctx = ctxlog.NewContextWithRecorder(ctx, "ext-secret-reconciler", mgr.GetEventRecorderFor("ext-secret-recorder"))
	log := ctxlog.ExtractLogger(ctx)
	r := NewExtendedSecretReconciler(ctx, config, mgr, credsgen.NewInMemoryGenerator(log), controllerutil.SetControllerReference)

	// Create a new controller
	c, err := controller.New("extendedsecret-controller", mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: config.MaxExtendedSecretWorkers,
	})
	if err != nil {
		return errors.Wrap(err, "Adding extended secret controller to manager failed.")
	}

	// Watch for changes to ExtendedSecrets
	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			o := e.Object.(*esv1.ExtendedSecret)
			secrets, err := listSecrets(ctx, mgr.GetClient(), o)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to list StatefulSets owned by ExtendedStatefulSet '%s': %s in extendedsecret reconciler", o.Name, err)
			}

			return len(secrets) == 0
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			o := e.ObjectOld.(*esv1.ExtendedSecret)
			n := e.ObjectNew.(*esv1.ExtendedSecret)
			if !reflect.DeepEqual(o.Spec, n.Spec) {
				ctxlog.NewPredicateEvent(e.ObjectNew).Debug(
					ctx, e.MetaNew, "esv1.ExtendedSecret",
					fmt.Sprintf("Update predicate passed for '%s'", e.MetaNew.GetName()),
				)
				return true
			}
			return false
		},
	}
	err = c.Watch(&source.Kind{Type: &esv1.ExtendedSecret{}}, &handler.EnqueueRequestForObject{}, p)
	if err != nil {
		return errors.Wrapf(err, "Watching extended secrets failed in extendedsecret controller.")
	}

	return nil
}

// listSecrets gets all Secrets owned by the ExtendedSecret
func listSecrets(ctx context.Context, client crc.Client, exSecret *esv1.ExtendedSecret) ([]corev1.Secret, error) {
	ctxlog.Debug(ctx, "Listing Secrets owned by ExtendedSecret '", exSecret.Name, "'.")

	secretLabels := map[string]string{esv1.LabelKind: esv1.GeneratedSecretKind}
	result := []corev1.Secret{}

	allSecrets := &corev1.SecretList{}
	err := client.List(ctx, allSecrets,
		crc.InNamespace(exSecret.Namespace),
		crc.MatchingLabels(secretLabels),
	)
	if err != nil {
		return nil, err
	}

	for _, secret := range allSecrets.Items {
		if metav1.IsControlledBy(&secret, exSecret) {
			result = append(result, secret)
			ctxlog.Debug(ctx, "Found Secret '", secret.Name, "' owned by ExtendedSecret '", exSecret.Name, "'.")
		}
	}

	if len(result) == 0 {
		ctxlog.Debug(ctx, "Did not find any Secret owned by ExtendedSecret '", exSecret.Name, "'.")
	}

	return result, nil
}
