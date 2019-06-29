package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

// DrainSafeReconciler reconciles a DrainSafe object
type DrainSafeReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=,resources=events,verbs=get;list;watch;create;update;patch;delete
func (r *DrainSafeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("drainsafe", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *DrainSafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		Complete(r)
}
