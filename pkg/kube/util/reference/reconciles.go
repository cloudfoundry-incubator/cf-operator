package reference

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"code.cloudfoundry.org/cf-operator/pkg/kube/apis"
	bdv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	ejobv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedjob/v1alpha1"
	estsv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedstatefulset/v1alpha1"
	log "code.cloudfoundry.org/cf-operator/pkg/kube/util/ctxlog"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/names"
	vss "code.cloudfoundry.org/cf-operator/pkg/kube/util/versionedsecretstore"
)

// ReconcileType lists all the types of reconciliations we can return,
// for controllers that have types that can reference ConfigMaps or Secrets
type ReconcileType int

const (
	// ReconcileForBOSHDeployment represents the BOSHDeployment CRD
	ReconcileForBOSHDeployment ReconcileType = iota
	// ReconcileForExtendedJob represents the ExtendedJob CRD
	ReconcileForExtendedJob
	// ReconcileForExtendedStatefulSet represents the ExtendedStatefulSet CRD
	ReconcileForExtendedStatefulSet
)

func (r ReconcileType) String() string {
	return [...]string{
		"BOSHDeployment",
		"ExtendedJob",
		"ExtendedStatefulSet",
	}[r]
}

// GetReconciles returns reconciliation requests for the BOSHDeployments, ExtendedJobs or ExtendedStatefulSets
// that reference an object. The object can be a ConfigMap or a Secret
func GetReconciles(ctx context.Context, client crc.Client, reconcileType ReconcileType, object apis.Object) ([]reconcile.Request, error) {
	isReferenceFor := func(parent interface{}) (bool, error) {
		var objectReferences map[string]bool
		var err error
		var name string

		switch object := object.(type) {
		case *corev1.ConfigMap:
			objectReferences, err = GetConfigMapsReferencedBy(parent)
			name = object.Name
		case *corev1.Secret:
			objectReferences, err = GetSecretsReferencedBy(parent)
			name = object.Name
			// When secret is versioned secret, we should check out versioned secret reference
			if isVersionedSecret(object) {
				return isVersionedSecretReference(objectReferences, name), nil
			}
		default:
			return false, errors.New("can't get reconciles for unknown object type; supported types are ConfigMap and Secret")
		}

		if err != nil {
			return false, errors.Wrap(err, "error listing references")
		}

		_, ok := objectReferences[name]

		return ok, nil
	}

	namespace := object.GetNamespace()
	result := []reconcile.Request{}

	switch reconcileType {
	case ReconcileForBOSHDeployment:
		log.Debugf(ctx, "Listing BOSHDeployments for object '%s' in namespace '%s'", object.GetName(), namespace)
		boshDeployments, err := listBOSHDeployments(ctx, client, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list BOSHDeployments for ConfigMap reconciles")
		}

		for _, boshDeployment := range boshDeployments.Items {
			isRef, err := isReferenceFor(boshDeployment)
			if err != nil {
				return nil, err
			}

			if isRef {
				result = append(result, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      boshDeployment.Name,
						Namespace: boshDeployment.Namespace,
					}})
			}
		}
	case ReconcileForExtendedJob:
		log.Debugf(ctx, "Listing ExtendedJobs for object '%s' in namespace '%s'", object.GetName(), namespace)
		extendedJobs, err := listExtendedJobs(ctx, client, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list ExtendedJobs for ConfigMap reconciles")
		}

		for _, eJob := range extendedJobs.Items {
			if !(eJob.Spec.UpdateOnConfigChange && eJob.IsAutoErrand()) {
				continue
			}
			isRef, err := isReferenceFor(eJob)
			if err != nil {
				return nil, err
			}

			if isRef {
				result = append(result, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      eJob.Name,
						Namespace: eJob.Namespace,
					}})
			}
		}
	case ReconcileForExtendedStatefulSet:
		log.Debugf(ctx, "Listing ExtendedStatefulSets for object '%s' in namespace '%s'", object.GetName(), namespace)
		extendedStatefulSets, err := listExtendedStatefulSets(ctx, client, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list ExtendedStatefulSets for ConfigMap reconciles")
		}

		for _, extendedStatefulSet := range extendedStatefulSets.Items {
			isRef, err := isReferenceFor(extendedStatefulSet)
			if err != nil {
				return nil, err
			}

			if isRef {
				result = append(result, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      extendedStatefulSet.Name,
						Namespace: extendedStatefulSet.Namespace,
					}})
			}
		}
	default:
		return nil, fmt.Errorf("unkown reconcile type %s", reconcileType.String())
	}

	return result, nil
}

// SkipReconciles returns true if the object is stale, and shouldn't be enqueued for reconciliation
// The object can be a ConfigMap or a Secret
func SkipReconciles(ctx context.Context, client crc.Client, object apis.Object) bool {
	var newResourceVersion string

	switch object := object.(type) {
	case *corev1.ConfigMap:
		cm := &corev1.ConfigMap{}
		err := client.Get(ctx, types.NamespacedName{Name: object.Name, Namespace: object.Namespace}, cm)
		if err != nil {
			log.Errorf(ctx, "Failed to get ConfigMap '%s': %s", object.Name, err)
			return true
		}

		newResourceVersion = cm.ResourceVersion
	case *corev1.Secret:
		s := &corev1.Secret{}
		err := client.Get(ctx, types.NamespacedName{Name: object.Name, Namespace: object.Namespace}, s)
		if err != nil {
			log.Errorf(ctx, "Failed to get Secret '%s': %s", object.Name, err)
			return true
		}

		newResourceVersion = s.ResourceVersion
	default:
		return false
	}

	if object.GetResourceVersion() != newResourceVersion {
		log.Debugf(ctx, "skip reconcile request for old resource version of '%s'", object.GetName())
		return true
	}
	return false
}

func listBOSHDeployments(ctx context.Context, client crc.Client, namespace string) (*bdv1.BOSHDeploymentList, error) {
	log.Debugf(ctx, "Listing BOSHDeployments in namespace '%s'", namespace)
	result := &bdv1.BOSHDeploymentList{}
	err := client.List(ctx, &crc.ListOptions{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list BOSHDeployments")
	}

	return result, nil
}

func listExtendedStatefulSets(ctx context.Context, client crc.Client, namespace string) (*estsv1.ExtendedStatefulSetList, error) {
	log.Debugf(ctx, "Listing ExtendedStatefulSets in namespace '%s'", namespace)
	result := &estsv1.ExtendedStatefulSetList{}
	err := client.List(ctx, &crc.ListOptions{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list ExtendedStatefulSets")
	}

	return result, nil
}

func listExtendedJobs(ctx context.Context, client crc.Client, namespace string) (*ejobv1.ExtendedJobList, error) {
	log.Debugf(ctx, "Listing ExtendedJobs in namespace '%s'", namespace)
	result := &ejobv1.ExtendedJobList{}
	err := client.List(ctx, &crc.ListOptions{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list ExtendedJobs")
	}

	return result, nil
}

func isVersionedSecret(secret *corev1.Secret) bool {
	labels := secret.Labels
	if labels == nil {
		return false
	}
	secretKind, ok := labels[vss.LabelSecretKind]
	if ok && secretKind == vss.VersionSecretKind {
		return true
	}

	return false
}

func isVersionedSecretReference(objectReferences map[string]bool, secretName string) bool {
	secretPrefix := names.GetPrefixFromVersionedSecretName(secretName)

	// The versioned secret reference can be original secret prefix (e.g. <name-prefix>)
	if len(secretPrefix) != 0 {
		_, ok := objectReferences[secretPrefix]
		if ok {
			return true
		}
	}

	// The versioned secret reference can also be secret prefix with version (e.g. <name-prefix>-v2)
	for ref := range objectReferences {
		refPrefix := names.GetPrefixFromVersionedSecretName(ref)
		if refPrefix == secretPrefix {
			return true
		}
	}

	return false
}
