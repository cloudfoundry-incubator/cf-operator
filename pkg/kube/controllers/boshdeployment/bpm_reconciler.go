package boshdeployment

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	qjv1a1 "code.cloudfoundry.org/quarks-job/pkg/kube/apis/quarksjob/v1alpha1"
	"code.cloudfoundry.org/quarks-operator/pkg/bosh/bpm"
	"code.cloudfoundry.org/quarks-operator/pkg/bosh/bpmconverter"
	bdm "code.cloudfoundry.org/quarks-operator/pkg/bosh/manifest"
	bdv1 "code.cloudfoundry.org/quarks-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	qstsv1a1 "code.cloudfoundry.org/quarks-operator/pkg/kube/apis/quarksstatefulset/v1alpha1"
	qstscontroller "code.cloudfoundry.org/quarks-operator/pkg/kube/controllers/quarksstatefulset"
	"code.cloudfoundry.org/quarks-operator/pkg/kube/util/boshdns"
	"code.cloudfoundry.org/quarks-operator/pkg/kube/util/mutate"
	"code.cloudfoundry.org/quarks-operator/pkg/kube/util/names"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	log "code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	"code.cloudfoundry.org/quarks-utils/pkg/meltdown"
	"code.cloudfoundry.org/quarks-utils/pkg/versionedsecretstore"
)

// BPMConverter converts k8s resources from single BOSH manifest
type BPMConverter interface {
	Resources(namespace string, manifestName string, dns bpmconverter.DNSSettings, qStsVersion string, instanceGroup *bdm.InstanceGroup, releaseImageProvider bdm.ReleaseImageProvider, bpmConfigs bpm.Configs, igResolvedSecretVersion string) (*bpmconverter.Resources, error)
}

// DesiredManifest unmarshals desired manifest from the manifest secret
type DesiredManifest interface {
	DesiredManifest(ctx context.Context, namespace string) (*bdm.Manifest, error)
}

// NewDNSFunc returns a dns client for the manifest
type NewDNSFunc func(m bdm.Manifest) (boshdns.DomainNameService, error)

var _ reconcile.Reconciler = &ReconcileBOSHDeployment{}

// NewBPMReconciler returns a new reconcile.Reconciler
func NewBPMReconciler(ctx context.Context, config *config.Config, mgr manager.Manager, resolver DesiredManifest, srf setReferenceFunc, converter BPMConverter, dns NewDNSFunc) reconcile.Reconciler {
	return &ReconcileBPM{
		ctx:                  ctx,
		config:               config,
		client:               mgr.GetClient(),
		scheme:               mgr.GetScheme(),
		resolver:             resolver,
		setReference:         srf,
		converter:            converter,
		versionedSecretStore: versionedsecretstore.NewVersionedSecretStore(mgr.GetClient()),
		newDNSFunc:           dns,
	}
}

// ReconcileBPM reconciles an Instance Group BPM versioned secret
type ReconcileBPM struct {
	ctx                  context.Context
	config               *config.Config
	client               client.Client
	scheme               *runtime.Scheme
	resolver             DesiredManifest
	setReference         setReferenceFunc
	converter            BPMConverter
	versionedSecretStore versionedsecretstore.VersionedSecretStore
	newDNSFunc           NewDNSFunc
}

// Reconcile reconciles an Instance Group BPM versioned secret read the corresponding
// desired manifest. It then applies BPM information and deploys instance groups.
func (r *ReconcileBPM) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	log.Infof(ctx, "Reconciling Instance Group BPM versioned secret '%s'", request.NamespacedName)
	bpmSecret := &corev1.Secret{}
	err := r.client.Get(ctx, request.NamespacedName, bpmSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Debug(ctx, "Skip reconcile: Instance Group BPM versioned secret not found")
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.WithEvent(bpmSecret, "GetBPMSecret").Errorf(ctx, "Failed to get Instance Group BPM versioned secret '%s': %v", request.NamespacedName, err)
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	if meltdown.NewAnnotationWindow(r.config.MeltdownDuration, bpmSecret.ObjectMeta.Annotations).Contains(time.Now()) {
		log.WithEvent(bpmSecret, "Meltdown").Debugf(ctx, "Resource '%s/%s' is in meltdown, requeue reconcile after %s", bpmSecret.Namespace, bpmSecret.Name, r.config.MeltdownRequeueAfter)
		return reconcile.Result{RequeueAfter: r.config.MeltdownRequeueAfter}, nil
	}

	// Get the labels from the BPM Secret and read the corresponding desired manifest
	deploymentName, ok := bpmSecret.Labels[bdv1.LabelDeploymentName]
	if !ok {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "LabelMissingError").Errorf(ctx, "There's no label for a BoshDeployment name on the BPM secret '%s'", request.NamespacedName)
	}

	instanceGroupName, ok := bpmSecret.Labels[qjv1a1.LabelRemoteID]
	if !ok {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "LabelMissingError").Errorf(ctx, "There's no label for a instance group name on the BPM secret '%s'", request.NamespacedName)
	}

	manifest, err := r.resolver.DesiredManifest(ctx, request.Namespace)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "DesiredManifestReadError").Errorf(ctx, "Failed to read desired manifest for bpm '%s': %v", request.NamespacedName, err)
	}

	dns, err := r.newDNSFunc(*manifest)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "DesiredManifestReadError").Errorf(ctx, "Failed to load BOSH DNS for manifest '%s': %v", deploymentName, err)
	}

	bdpl := &bdv1.BOSHDeployment{}
	err = r.client.Get(ctx, types.NamespacedName{Namespace: request.Namespace, Name: deploymentName}, bdpl)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "GetBOSHDeployment").Errorf(ctx, "Failed to get BoshDeployment instance '%s/%s': %v", request.Namespace, deploymentName, err)
	}

	err = dns.Apply(ctx, request.Namespace, r.client, func(object metav1.Object) error {
		return r.setReference(bdpl, object, r.scheme)
	})

	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "DnsReconcileError").Errorf(ctx, "Failed to reconcile dns: %v", err)
	}

	// Apply BPM information
	resources, err := r.applyBPMResources(bdpl.Name, instanceGroupName, bpmSecret, manifest, dns)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "BPMApplyingError").Errorf(ctx, "Failed to apply BPM information: %v", err)
	}

	if resources == nil {
		log.WithEvent(bpmSecret, "SkipReconcile").Infof(ctx, "Skip reconcile: BPM resources not found")
		return reconcile.Result{}, nil
	}

	// Deploy instance groups
	err = r.deployInstanceGroups(ctx, bdpl, instanceGroupName, resources)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "InstanceGroupStartError").Errorf(ctx, "Failed to start: %v", err)
	}

	meltdown.SetLastReconcile(&bpmSecret.ObjectMeta, time.Now())
	err = r.client.Update(ctx, bpmSecret)
	if err != nil {
		log.WithEvent(bpmSecret, "UpdateError").Errorf(ctx, "Failed to update reconcile timestamp on BPM versioned secret '%s' (%v): %s", bpmSecret.Name, bpmSecret.ResourceVersion, err)
		return reconcile.Result{Requeue: false}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBPM) applyBPMResources(bdplName string, instanceGroupName string, bpmSecret *corev1.Secret, manifest *bdm.Manifest, dns boshdns.DomainNameService) (*bpmconverter.Resources, error) {
	var bpmInfo bdm.BPMInfo
	if val, ok := bpmSecret.Data["bpm.yaml"]; ok {
		err := yaml.Unmarshal(val, &bpmInfo)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Couldn't find bpm.yaml key in manifest secret")
	}

	instanceGroup, found := manifest.InstanceGroups.InstanceGroupByName(instanceGroupName)
	if !found {
		return nil, errors.Errorf("instance group '%s' not found", instanceGroupName)
	}

	// Fetch qSts version
	quarksStatefulSet := &qstsv1a1.QuarksStatefulSet{}
	quarksStatefulSetName := instanceGroup.NameSanitized()
	err := r.client.Get(r.ctx, types.NamespacedName{Namespace: bpmSecret.Namespace, Name: quarksStatefulSetName}, quarksStatefulSet)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, errors.Errorf("Failed to get QuarksStatefulSet instance '%s/%s': %v", bpmSecret.Namespace, quarksStatefulSetName, err)
		}
	}

	qStsVersionString := "0"
	if quarksStatefulSet.Namespace != "" {
		_, qStsVersion, err := qstscontroller.GetMaxStatefulSetVersion(r.ctx, r.client, quarksStatefulSet)
		if err != nil {
			return nil, err
		}
		qStsVersion = qStsVersion + 1
		qStsVersionString = strconv.Itoa(qStsVersion)
	}

	igResolvedSecretVersion, err := r.fetchIGresolvedVersion(bpmSecret.Namespace, instanceGroupName)
	if err != nil {
		return nil, err
	}

	resources, err := r.converter.Resources(bpmSecret.Namespace, bdplName, dns, qStsVersionString, instanceGroup, manifest, bpmInfo.Configs, igResolvedSecretVersion)
	if err != nil {
		return resources, err
	}

	return resources, nil
}

func (r *ReconcileBPM) fetchIGresolvedVersion(namespace string, instanceGroupName string) (string, error) {
	igResolvedSecretName := names.InstanceGroupSecretName(instanceGroupName, "")
	igResolvedSecret, err := r.versionedSecretStore.Latest(r.ctx, namespace, igResolvedSecretName)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read latest versioned secret '%s/%s'", namespace, igResolvedSecretName)
	}
	return igResolvedSecret.GetLabels()[versionedsecretstore.LabelVersion], nil
}

// deployInstanceGroups create or update QuarksJobs and QuarksStatefulSets for instance groups
func (r *ReconcileBPM) deployInstanceGroups(ctx context.Context, bdpl *bdv1.BOSHDeployment, instanceGroupName string, resources *bpmconverter.Resources) error {
	log.Debugf(ctx, "Creating quarksJobs and quarksStatefulSets for instance group '%s'", instanceGroupName)

	for _, qJob := range resources.Errands {
		if qJob.Labels[bdv1.LabelInstanceGroupName] != instanceGroupName {
			log.Debugf(ctx, "Skipping apply QuarksJob '%s/%s' for instance group '%s' because of mismatching '%s' label", bdpl.Namespace, qJob.Name, bdpl.Name, bdv1.LabelInstanceGroupName)
			continue
		}

		if err := r.setReference(bdpl, &qJob, r.scheme); err != nil {
			return log.WithEvent(bdpl, "QuarksJobForDeploymentError").Errorf(ctx, "Failed to set reference for QuarksJob instance group '%s' : %v", instanceGroupName, err)
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.client, &qJob, mutate.QuarksJobMutateFn(&qJob))
		if err != nil {
			return log.WithEvent(bdpl, "ApplyQuarksJobError").Errorf(ctx, "Failed to apply QuarksJob for instance group '%s' : %v", instanceGroupName, err)
		}

		log.Debugf(ctx, "QuarksJob '%s/%s' has been %s", bdpl.Namespace, qJob.Name, op)
	}

	for _, svc := range resources.Services {
		if svc.Labels[bdv1.LabelInstanceGroupName] != instanceGroupName {
			log.Debugf(ctx, "Skipping apply Service '%s/%s' for instance group '%s' because of mismatching '%s' label", bdpl.Namespace, svc.Name, bdpl.Name, bdv1.LabelInstanceGroupName)
			continue
		}

		if err := r.setReference(bdpl, &svc, r.scheme); err != nil {
			return log.WithEvent(bdpl, "ServiceForDeploymentError").Errorf(ctx, "Failed to set reference for Service instance group '%s' : %v", instanceGroupName, err)
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.client, &svc, mutate.ServiceMutateFn(&svc))
		if err != nil {
			return log.WithEvent(bdpl, "ApplyServiceError").Errorf(ctx, "Failed to apply Service for instance group '%s' : %v", instanceGroupName, err)
		}

		log.Debugf(ctx, "Service '%s/%s' has been %s", bdpl.Namespace, svc.Name, op)
	}

	for _, qSts := range resources.InstanceGroups {
		if qSts.Labels[bdv1.LabelInstanceGroupName] != instanceGroupName {
			log.Debugf(ctx, "Skipping apply QuarksStatefulSet '%s/%s' for instance group '%s' because of mismatching '%s' label", bdpl.Namespace, qSts.Name, bdpl.Name, bdv1.LabelInstanceGroupName)
			continue
		}

		if err := r.setReference(bdpl, &qSts, r.scheme); err != nil {
			return log.WithEvent(bdpl, "QuarksStatefulSetForDeploymentError").Errorf(ctx, "Failed to set reference for QuarksStatefulSet instance group '%s' : %v", instanceGroupName, err)
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.client, &qSts, mutate.QuarksStatefulSetMutateFn(&qSts))
		if err != nil {
			return log.WithEvent(bdpl, "ApplyQuarksStatefulSetError").Errorf(ctx, "Failed to apply QuarksStatefulSet for instance group '%s' : %v", instanceGroupName, err)
		}

		log.Debugf(ctx, "QuarksStatefulSet '%s/%s' has been %s", bdpl.Namespace, qSts.Name, op)
	}

	return nil
}
