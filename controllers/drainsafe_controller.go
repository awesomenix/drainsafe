package controllers

import (
	"context"
	"os"
	"time"

	"github.com/awesomenix/drainsafe/events"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
)

// DrainSafeReconciler reconciles a DrainSafe object
type DrainSafeReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile consumes event
func (r *DrainSafeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("drainsafe", req.NamespacedName)

	// your logic here

	event := &corev1.Event{}
	err := r.Get(ctx, req.NamespacedName, event)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "faled to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	if event.Namespace != os.Getenv("POD_NAMESPACE") {
		return ctrl.Result{}, nil
	}

	log.Info("got event",
		"Name", event.Name,
		"Namespace", event.Namespace,
		"Reason", event.Reason,
		"Message", event.Message,
		"Type", event.Type,
		"Timestamp", event.LastTimestamp)

	if event.Reason == events.Scheduled {
		err = Cordon(event.Message)
		if err != nil {
			log.Error(err, "failed to cordon vm", "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(event, "Normal", "Cordoned", "%s", event.Message)
	}

	if event.Reason == events.Cordoned {
		err = Drain(event.Message)
		if err != nil {
			log.Error(err, "failed to cordon vm", "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(event, "Normal", "Drained", "%s", event.Message)
	}

	if event.Reason == events.Running {
		err = Uncordon(event.Message)
		if err != nil {
			log.Error(err, "failed to cordon vm", "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(event, "Normal", "Uncordoned", "%s", event.Message)
	}

	return ctrl.Result{}, nil
}

func (r *DrainSafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		Complete(r)
}
