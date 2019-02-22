package extendedjob

import (
	ejv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedjob/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/context"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddErrand creates a new ExtendedJob controller and adds it to the Manager
func AddErrand(log *zap.SugaredLogger, ctrConfig *context.Config, mgr manager.Manager) error {
	f := controllerutil.SetControllerReference
	r := NewErrandReconciler(log, ctrConfig, mgr, f)
	c, err := controller.New("extendedjob-errand-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Only trigger if Spec.Run is 'now'
	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			exJob := e.Object.(*ejv1.ExtendedJob)
			return exJob.Spec.Trigger.Strategy == ejv1.TriggerNow || exJob.Spec.Trigger.Strategy == ejv1.TriggerOnce
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldExJob := e.ObjectOld.(*ejv1.ExtendedJob)
			newExJob := e.ObjectNew.(*ejv1.ExtendedJob)
			run := newExJob.Spec.Trigger.Strategy == ejv1.TriggerNow && oldExJob.Spec.Trigger.Strategy == ejv1.TriggerManually
			return run
		},
	}
	err = c.Watch(&source.Kind{Type: &ejv1.ExtendedJob{}}, &handler.EnqueueRequestForObject{}, p)
	return err
}
