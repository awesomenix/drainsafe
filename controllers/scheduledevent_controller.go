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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

// DrainSafeReconciler reconciles a DrainSafe object
type ScheduledEventReconciler struct {
	client.Client
	Log            logr.Logger
	Recorder       record.EventRecorder
	pod            *corev1.Pod
	StopCh         <-chan struct{}
	hostname       string
	vmInstanceName string
}

// Reconcile consumes event
func (r *ScheduledEventReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("scheduledevent", req.NamespacedName)

	// your logic here

	event := &corev1.Event{}
	err := r.Get(ctx, req.NamespacedName, event)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	r.logRunning()

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

	if event.Reason == events.Drained {
		err = approveScheduledEvent(r.vmInstanceName)
		if err != nil {
			log.Error(err, "failed to approve scheduled event")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		if r.pod != nil {
			r.Recorder.Eventf(r.pod, "Normal", events.Started, "%s", r.hostname)
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *ScheduledEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.startup(); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Event{}).
		Complete(r)
}

func (r *ScheduledEventReconciler) startup() error {
	hostname := os.Getenv("NODE_NAME")

	vmInstanceName, err := getVMInstanceName()
	if err != nil {
		log.Error(err, "failed to get vm instance name")
		return err
	}

	r.hostname = hostname
	r.vmInstanceName = vmInstanceName

	go r.eventWatcher()

	return nil
}

func (r *ScheduledEventReconciler) eventWatcher() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.logRunning()
			isScheduled, err := isScheduledEvent(r.vmInstanceName)
			if err != nil {
				log.Error(err, "failed to find scheduled events")
				continue
			}
			if isScheduled {
				r.Recorder.Eventf(r.pod, "Normal", events.Scheduled, "%s", r.hostname)
			}
		case <-r.StopCh:
			return
		}
	}
}

func (r *ScheduledEventReconciler) logRunning() {
	if r.pod != nil {
		return
	}
	namespacedName := types.NamespacedName{
		Name:      os.Getenv("POD_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}

	pod := &corev1.Pod{}
	if err := r.Get(context.TODO(), namespacedName, pod); err != nil {
		log.Error(err, "failed to get pod", "NamespacedName", namespacedName.String())
		return
	}
	r.pod = pod
	r.Recorder.Eventf(r.pod, "Normal", events.Running, "%s", r.hostname)
}
