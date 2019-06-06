package extendedstatefulset

import (
	"context"
	"reflect"

	"code.cloudfoundry.org/cf-operator/pkg/kube/util/reference"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	estsv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedstatefulset/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/config"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/ctxlog"
)

// Add creates a new ExtendedStatefulSet controller and adds it to the Manager
func Add(ctx context.Context, config *config.Config, mgr manager.Manager) error {
	ctx = ctxlog.NewContextWithRecorder(ctx, "ext-statefulset-reconciler", mgr.GetRecorder("ext-statefulset-recorder"))
	r := NewReconciler(ctx, config, mgr, controllerutil.SetControllerReference)

	// Create a new controller
	c, err := controller.New("extendedstatefulset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ExtendedStatefulSet
	err = c.Watch(&source.Kind{Type: &estsv1.ExtendedStatefulSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch ConfigMaps referenced by the ExtendedStatefulSet
	configMapPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			configMap := e.Object.(*corev1.ConfigMap)

			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, configMap)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for configMap '%s': %v", configMap.Name, err)
			}

			// The ConfigMap should reference at least one ExtendedStatefulSet in order for us to consider it
			return len(reconciles) > 0
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConfigMap := e.ObjectOld.(*corev1.ConfigMap)
			newConfigMap := e.ObjectNew.(*corev1.ConfigMap)

			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, newConfigMap)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for configMap '%s': %v", newConfigMap.Name, err)
			}

			// The ConfigMap should reference at least one ExtendedStatefulSet in order for us to consider it
			return len(reconciles) > 0 && !reflect.DeepEqual(oldConfigMap.Data, newConfigMap.Data)
		},
	}
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			config := a.Object.(*corev1.ConfigMap)

			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, config)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for configMap '%s': %v", config.Name, err)
			}

			return reconciles
		}),
	}, configMapPredicates)
	if err != nil {
		return err
	}

	// Watch Secrets referenced by the ExtendedStatefulSet
	secretPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			secret := e.Object.(*corev1.Secret)
			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, secret)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for secret '%s': %v", secret.Name, err)
			}

			// The Secret should reference at least one ExtendedStatefulSet in order for us to consider it
			return len(reconciles) > 1
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSecret := e.ObjectOld.(*corev1.Secret)
			newSecret := e.ObjectNew.(*corev1.Secret)

			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, newSecret)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for secret '%s': %v", newSecret.Name, err)
			}

			return len(reconciles) > 1 && !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		},
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			secret := a.Object.(*corev1.Secret)

			reconciles, err := reference.GetReconciles(ctx, mgr.GetClient(), reference.ReconcileForExtendedStatefulSet, secret)
			if err != nil {
				ctxlog.Errorf(ctx, "Failed to calculate reconciles for secret '%s': %v", secret.Name, err)
			}

			return reconciles
		}),
	}, secretPredicates)
	if err != nil {
		return err
	}

	return nil
}
