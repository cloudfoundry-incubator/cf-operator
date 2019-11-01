package extendedstatefulset

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	estsv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedstatefulset/v1alpha1"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	"code.cloudfoundry.org/quarks-utils/pkg/meltdown"
	vss "code.cloudfoundry.org/quarks-utils/pkg/versionedsecretstore"
)

const (
	// EnvKubeAz is set by available zone name
	EnvKubeAz = "KUBE_AZ"
	// EnvBoshAz is set by available zone name
	EnvBoshAz = "BOSH_AZ"
	// EnvReplicas describes the number of replicas in the ExtendedStatefulSet
	EnvReplicas = "REPLICAS"
	// EnvCfOperatorAz is set by available zone name
	EnvCfOperatorAz = "CF_OPERATOR_AZ"
	// EnvCFOperatorAZIndex is set by available zone index
	EnvCFOperatorAZIndex = "AZ_INDEX"
)

// Check that ReconcileExtendedStatefulSet implements the reconcile.Reconciler interface
var _ reconcile.Reconciler = &ReconcileExtendedStatefulSet{}

type setReferenceFunc func(owner, object metav1.Object, scheme *runtime.Scheme) error

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(ctx context.Context, config *config.Config, mgr manager.Manager, srf setReferenceFunc, store vss.VersionedSecretStore) reconcile.Reconciler {
	return &ReconcileExtendedStatefulSet{
		ctx:                  ctx,
		config:               config,
		client:               mgr.GetClient(),
		scheme:               mgr.GetScheme(),
		setReference:         srf,
		versionedSecretStore: store,
	}
}

// ReconcileExtendedStatefulSet reconciles an ExtendedStatefulSet object
type ReconcileExtendedStatefulSet struct {
	ctx                  context.Context
	client               client.Client
	scheme               *runtime.Scheme
	setReference         setReferenceFunc
	config               *config.Config
	versionedSecretStore vss.VersionedSecretStore
}

// Reconcile reads that state of the cluster for a ExtendedStatefulSet object
// and makes changes based on the state read and what is in the ExtendedStatefulSet.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileExtendedStatefulSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Fetch the ExtendedStatefulSet we need to reconcile
	exStatefulSet := &estsv1.ExtendedStatefulSet{}

	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	ctxlog.Info(ctx, "Reconciling ExtendedStatefulSet ", request.NamespacedName)
	err := r.client.Get(ctx, request.NamespacedName, exStatefulSet)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			ctxlog.Debug(ctx, "Skip ExtendedStatefulset reconcile: ExtendedStatefulSet not found")
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if meltdown.NewWindow(r.config.MeltdownDuration, exStatefulSet.Status.LastReconcile).Contains(time.Now()) {
		ctxlog.WithEvent(exStatefulSet, "Meltdown").Debugf(ctx, "Resource '%s' is in meltdown, requeue reconcile after %s", exStatefulSet.Name, r.config.MeltdownRequeueAfter)
		return reconcile.Result{RequeueAfter: r.config.MeltdownRequeueAfter}, nil
	}

	// Get the current StatefulSet.
	currentStatefulSet, currentVersion, err := r.getCurrentStatefulSet(ctx, exStatefulSet)
	if err != nil {
		return reconcile.Result{}, ctxlog.WithEvent(exStatefulSet, "StatefulSetNotFound").Error(ctx, "Could not retrieve latest StatefulSet owned by ExtendedStatefulSet '", request.NamespacedName, "': ", err)
	}

	// Calculate the desired statefulSets
	desiredStatefulSets, desiredVersion, err := r.calculateDesiredStatefulSets(exStatefulSet, currentVersion)
	if err != nil {
		return reconcile.Result{}, ctxlog.WithEvent(exStatefulSet, "CalculationError").Error(ctx, "Could not calculate StatefulSet owned by ExtendedStatefulSet '", request.NamespacedName, "': ", err)
	}

	if exStatefulSet.Spec.Template.Spec.VolumeClaimTemplates != nil {
		err := r.alterVolumeManagementStatefulSet(ctx, currentVersion, desiredVersion, exStatefulSet, currentStatefulSet)
		if err != nil {
			ctxlog.Error(ctx, "Alteration of VolumeManagement statefulset failed for ExtendedStatefulset ", exStatefulSet.Name, " in namespace ", exStatefulSet.Namespace, ".", err)
			return reconcile.Result{}, err
		}
	}

	for _, desiredStatefulSet := range desiredStatefulSets {
		desiredStatefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{}

		// If it doesn't exist, create it
		ctxlog.Info(ctx, "StatefulSet '", desiredStatefulSet.Name, "' owned by ExtendedStatefulSet '", request.NamespacedName, "' not found, will be created.")

		r.versionedSecretStore.SetSecretReferences(ctx, request.Namespace, &exStatefulSet.Spec.Template.Spec.Template.Spec)

		if err := r.createStatefulSet(ctx, exStatefulSet, &desiredStatefulSet); err != nil {
			return reconcile.Result{}, ctxlog.WithEvent(exStatefulSet, "CreateStatefulSetError").Error(ctx, "Could not create StatefulSet for ExtendedStatefulSet '", request.NamespacedName, "': ", err)
		}
	}

	now := metav1.Now()
	exStatefulSet.Status.LastReconcile = &now
	err = r.client.Status().Update(ctx, exStatefulSet)
	if err != nil {
		ctxlog.WithEvent(exStatefulSet, "UpdateStatusError").Errorf(ctx, "Failed to update reconcile timestamp on ExtendedStatefulSet '%s' (%v): %s", exStatefulSet.Name, exStatefulSet.ResourceVersion, err)
		return reconcile.Result{Requeue: false}, nil
	}

	return reconcile.Result{}, nil
}

// calculateDesiredStatefulSets generates the desired StatefulSets that should exist
func (r *ReconcileExtendedStatefulSet) calculateDesiredStatefulSets(exStatefulSet *estsv1.ExtendedStatefulSet, currentVersion int) ([]v1beta2.StatefulSet, int, error) {
	var desiredStatefulSets []v1beta2.StatefulSet

	template := exStatefulSet.Spec.Template.DeepCopy()

	// Place the StatefulSet in the same namespace as the ExtendedStatefulSet
	template.SetNamespace(exStatefulSet.Namespace)

	if template.Annotations == nil {
		template.Annotations = map[string]string{}
	}

	// Set version
	desiredVersion := currentVersion + 1
	template.Annotations[estsv1.AnnotationVersion] = fmt.Sprintf("%d", desiredVersion)

	if exStatefulSet.Spec.ZoneNodeLabel == "" {
		exStatefulSet.Spec.ZoneNodeLabel = estsv1.DefaultZoneNodeLabel
	}

	if len(exStatefulSet.Spec.Zones) > 0 {
		for zoneIndex, zoneName := range exStatefulSet.Spec.Zones {
			statefulSet, err := r.generateSingleStatefulSet(exStatefulSet, template, zoneIndex, zoneName, desiredVersion)
			if err != nil {
				return desiredStatefulSets, desiredVersion, errors.Wrapf(err, "Could not generate StatefulSet template for AZ '%d/%s'", zoneIndex, zoneName)
			}
			desiredStatefulSets = append(desiredStatefulSets, *statefulSet)
		}

	} else {
		statefulSet, err := r.generateSingleStatefulSet(exStatefulSet, template, 0, "", desiredVersion)
		if err != nil {
			return desiredStatefulSets, desiredVersion, errors.Wrap(err, "Could not generate StatefulSet template for single zone")
		}
		desiredStatefulSets = append(desiredStatefulSets, *statefulSet)
	}

	return desiredStatefulSets, desiredVersion, nil
}

// createStatefulSet creates a StatefulSet
func (r *ReconcileExtendedStatefulSet) createStatefulSet(ctx context.Context, exStatefulSet *estsv1.ExtendedStatefulSet, statefulSet *v1beta2.StatefulSet) error {

	// Set the owner of the StatefulSet, so it's garbage collected,
	// and we can find it later
	ctxlog.Info(ctx, "Setting owner for StatefulSet '", statefulSet.Name, "' to ExtendedStatefulSet '", exStatefulSet.Name, "' in namespace '", exStatefulSet.Namespace, "'.")
	if err := r.setReference(exStatefulSet, statefulSet, r.scheme); err != nil {
		return errors.Wrapf(err, "could not set owner for StatefulSet '%s' to ExtendedStatefulSet '%s' in namespace '%s'", statefulSet.Name, exStatefulSet.Name, exStatefulSet.Namespace)
	}

	// Create the StatefulSet
	if err := r.client.Create(ctx, statefulSet); err != nil {
		return errors.Wrapf(err, "could not create StatefulSet '%s' for ExtendedStatefulSet '%s' in namespace '%s'", statefulSet.Name, exStatefulSet.Name, exStatefulSet.Namespace)
	}

	ctxlog.Info(ctx, "Created StatefulSet '", statefulSet.Name, "' for ExtendedStatefulSet '", exStatefulSet.Name, "' in namespace '", exStatefulSet.Namespace, "'.")

	return nil
}

// getCurrentStatefulSet gets the latest (by version) StatefulSet owned by the ExtendedStatefulSet
func (r *ReconcileExtendedStatefulSet) getCurrentStatefulSet(ctx context.Context, exStatefulSet *estsv1.ExtendedStatefulSet) (*v1beta2.StatefulSet, int, error) {
	// Default response is an empty StatefulSet with version '0' and an empty signature
	result := &v1beta2.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				estsv1.AnnotationVersion: "0",
			},
		},
	}
	maxVersion := 0

	// Get all owned StatefulSets
	statefulSets, err := listStatefulSetsFromInformer(ctx, r.client, exStatefulSet)
	if err != nil {
		return nil, 0, err
	}

	ctxlog.Debug(ctx, "Getting the latest StatefulSet owned by ExtendedStatefulSet '", exStatefulSet.Name, "'.")

	for _, ss := range statefulSets {
		strVersion := ss.Annotations[estsv1.AnnotationVersion]
		if strVersion == "" {
			return nil, 0, errors.Errorf("The statefulset %s does not have the annotation(%s), a version could not be retrieved.", ss.Name, estsv1.AnnotationVersion)
		}

		version, err := strconv.Atoi(strVersion)
		if err != nil {
			return nil, 0, err
		}

		if ss.Annotations != nil && version > maxVersion {
			result = &ss
			maxVersion = version
		}
	}

	return result, maxVersion, nil
}

// generateSingleStatefulSet creates a StatefulSet from one zone
func (r *ReconcileExtendedStatefulSet) generateSingleStatefulSet(exStatefulSet *estsv1.ExtendedStatefulSet, template *v1beta2.StatefulSet, zoneIndex int, zoneName string, version int) (*v1beta2.StatefulSet, error) {
	statefulSet := template.DeepCopy()

	// Get the labels and annotations
	labels := statefulSet.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	annotations := statefulSet.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	statefulSetNamePrefix := exStatefulSet.GetName()

	// Get the pod labels and annotations
	podLabels := statefulSet.Spec.Template.GetLabels()
	if podLabels == nil {
		podLabels = make(map[string]string)
	}
	podAnnotations := statefulSet.Spec.Template.GetAnnotations()
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}

	// Update available-zone specified properties
	if zoneName != "" {
		// Override name prefix with zoneIndex
		statefulSetNamePrefix = fmt.Sprintf("%s-z%d", exStatefulSet.GetName(), zoneIndex)

		labels[estsv1.LabelAZIndex] = strconv.Itoa(zoneIndex)
		labels[estsv1.LabelAZName] = zoneName

		zonesBytes, err := json.Marshal(exStatefulSet.Spec.Zones)
		if err != nil {
			return &v1beta2.StatefulSet{}, errors.Wrapf(err, "Could not marshal zones: '%v'", exStatefulSet.Spec.Zones)
		}
		annotations[estsv1.AnnotationZones] = string(zonesBytes)

		podLabels[estsv1.LabelAZIndex] = strconv.Itoa(zoneIndex)
		podLabels[estsv1.LabelAZName] = zoneName

		podAnnotations[estsv1.AnnotationZones] = string(zonesBytes)

		statefulSet = r.updateAffinity(statefulSet, exStatefulSet.Spec.ZoneNodeLabel, zoneName)
	}

	podLabels[estsv1.LabelAZIndex] = strconv.Itoa(zoneIndex)
	podLabels[estsv1.LabelEStsName] = exStatefulSet.GetName()

	statefulSet.Spec.Template.SetLabels(podLabels)
	statefulSet.Spec.Template.SetAnnotations(podAnnotations)

	r.injectContainerEnv(&statefulSet.Spec.Template.Spec, zoneIndex, zoneName, exStatefulSet.Spec.Template.Spec.Replicas)

	annotations[estsv1.AnnotationVersion] = fmt.Sprintf("%d", version)

	// Set updated properties
	statefulSet.SetName(fmt.Sprintf("%s-v%d", statefulSetNamePrefix, version))
	statefulSet.SetLabels(labels)
	statefulSet.SetAnnotations(annotations)

	return statefulSet, nil
}

// updateAffinity Update current statefulSet Affinity from AZ specification
func (r *ReconcileExtendedStatefulSet) updateAffinity(statefulSet *v1beta2.StatefulSet, zoneNodeLabel string, zoneName string) *v1beta2.StatefulSet {
	nodeInZoneSelector := corev1.NodeSelectorRequirement{
		Key:      zoneNodeLabel,
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{zoneName},
	}

	affinity := statefulSet.Spec.Template.Spec.Affinity
	// Check if optional properties were set
	if affinity == nil {
		affinity = &corev1.Affinity{}
	}

	if affinity.NodeAffinity == nil {
		affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						nodeInZoneSelector,
					},
				},
			},
		}
	} else {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				nodeInZoneSelector,
			},
		})
	}

	statefulSet.Spec.Template.Spec.Affinity = affinity

	return statefulSet
}

// injectContainerEnv inject AZ info to container envs
func (r *ReconcileExtendedStatefulSet) injectContainerEnv(podSpec *corev1.PodSpec, zoneIndex int, zoneName string, replicas *int32) {

	containers := []*corev1.Container{}
	for i := 0; i < len(podSpec.Containers); i++ {
		containers = append(containers, &podSpec.Containers[i])
	}
	for i := 0; i < len(podSpec.InitContainers); i++ {
		containers = append(containers, &podSpec.InitContainers[i])
	}
	for _, container := range containers {
		envs := container.Env

		if zoneIndex >= 0 {
			envs = upsertEnvs(envs, EnvKubeAz, zoneName)
			envs = upsertEnvs(envs, EnvBoshAz, zoneName)
			envs = upsertEnvs(envs, EnvCfOperatorAz, zoneName)
			envs = upsertEnvs(envs, EnvCFOperatorAZIndex, strconv.Itoa(zoneIndex+1))
		} else {
			// Default to zone 1
			envs = upsertEnvs(envs, EnvCFOperatorAZIndex, "1")
		}
		envs = upsertEnvs(envs, EnvReplicas, strconv.Itoa(int(*replicas)))

		container.Env = envs
	}
}

func upsertEnvs(envs []corev1.EnvVar, name string, value string) []corev1.EnvVar {
	for idx, env := range envs {
		if env.Name == name {
			envs[idx].Value = value
			return envs
		}
	}

	envs = append(envs, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return envs
}
