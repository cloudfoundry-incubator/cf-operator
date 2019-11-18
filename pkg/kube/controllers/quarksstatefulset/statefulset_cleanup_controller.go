package quarksstatefulset

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/api/apps/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	qstsv1a1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/quarksstatefulset/v1alpha1"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
)

// AddStatefulSetCleanup creates a new statefulset cleanup controller and adds it to the manager.
// The purpose of this controller is to delete the temporary statefulset used to keep the volumes alive.
func AddStatefulSetCleanup(ctx context.Context, config *config.Config, mgr manager.Manager) error {
	ctx = ctxlog.NewContextWithRecorder(ctx, "statefulset-cleanup-reconciler", mgr.GetEventRecorderFor("statefulset-cleanup-recorder"))
	r := NewStatefulSetCleanupReconciler(ctx, config, mgr)

	// Create a new controller
	c, err := controller.New("statefulset-cleanup-controller", mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: config.MaxQuarksStatefulSetWorkers,
	})
	if err != nil {
		return errors.Wrap(err, "Adding StatefulSet cleanup controller to manager failed.")
	}

	// Watch StatefulSets owned by the QuarksStatefulSet
	// Trigger when
	// - at least one pod of new version is running
	// - all pods of volume management are running
	statefulSetPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			newStatefulSet := e.Object.(*v1beta2.StatefulSet)
			enqueueForVolumeManagementStatefulSet := isVolumeManagementStatefulSet(newStatefulSet.Name) && newStatefulSet.Status.ReadyReplicas > 0 && newStatefulSet.Status.ReadyReplicas == newStatefulSet.Status.CurrentReplicas
			enqueueForVersionStatefulSet := newStatefulSet.Status.ReadyReplicas > 0

			if enqueueForVersionStatefulSet || enqueueForVolumeManagementStatefulSet {
				ctxlog.NewPredicateEvent(e.Object).Debug(
					ctx, e.Meta, "v1beta2.StatefulSet",
					fmt.Sprintf("Create predicate passed for '%s'", e.Meta.GetName()),
				)
				return true
			}
			return false
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			newStatefulSet := e.ObjectNew.(*v1beta2.StatefulSet)
			enqueueForVolumeManagementStatefulSet := isVolumeManagementStatefulSet(newStatefulSet.Name) && newStatefulSet.Status.ReadyReplicas > 0 && newStatefulSet.Status.ReadyReplicas == newStatefulSet.Status.CurrentReplicas
			enqueueForVersionStatefulSet := newStatefulSet.Status.ReadyReplicas > 0

			if enqueueForVersionStatefulSet || enqueueForVolumeManagementStatefulSet {
				ctxlog.NewPredicateEvent(e.ObjectNew).Debug(
					ctx, e.MetaNew, "v1beta2.StatefulSet",
					fmt.Sprintf("Update predicate passed for '%s'", e.MetaNew.GetName()),
				)
				return true
			}
			return false
		},
	}
	err = c.Watch(&source.Kind{Type: &v1beta2.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: false,
		OwnerType:    &qstsv1a1.QuarksStatefulSet{},
	}, statefulSetPredicates)
	if err != nil {
		return errors.Wrapf(err, "Watching statefulSet failed in statefulSet cleanup controller.")
	}

	return nil
}
