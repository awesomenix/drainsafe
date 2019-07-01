// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"os"
	"time"

	"github.com/awesomenix/drainsafe/annotations"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

// ScheduledEventReconciler reconciles a DrainSafe object
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

	node := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	if node.Annotations == nil {
		return ctrl.Result{}, nil
	}

	maintenance := node.Annotations[annotations.DrainSafeMaintenance]

	log.Info("got update event",
		"Name", node.Name,
		"Maintenance", maintenance,
		"Annotations", node.Annotations)

	if maintenance == annotations.Drained {
		err = approveScheduledEvent(r.vmInstanceName)
		if err != nil {
			log.Error(err, "failed to approve scheduled event")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		return r.updateNodeState(node, annotations.Started)
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *ScheduledEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.startup(); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
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
			node := &corev1.Node{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: r.hostname}, node)
			if err != nil {
				log.Error(err, "failed to get node", "Name", r.hostname)
				continue
			}
			maintenance := node.Annotations[annotations.DrainSafeMaintenance]
			isScheduled, err := isScheduledEvent(r.vmInstanceName)
			if err != nil {
				log.Error(err, "failed to find scheduled events")
				continue
			}
			if isScheduled {
				if maintenance != annotations.Running {
					log.Info("node is under going maintenance, skipping setting annotation", "Maintenance", maintenance)
					continue
				}
				r.updateNodeState(node, annotations.Scheduled)
				continue
			}
			r.updateNodeState(node, annotations.Running)
		case <-r.StopCh:
			return
		}
	}
}

func (r *ScheduledEventReconciler) updateNodeState(node *corev1.Node, state string) (ctrl.Result, error) {
	if node.Annotations[annotations.DrainSafeMaintenance] == state {
		return ctrl.Result{}, nil
	}
	log.Info("updating node state", "Current", node.Annotations[annotations.DrainSafeMaintenance], "Desired", state)
	node.Annotations[annotations.DrainSafeMaintenance] = state
	if err := r.Update(context.TODO(), node); err != nil {
		log.Error(err, "failed to update node")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.Recorder.Eventf(node, "Normal", state, "%s by %s", node.Name, os.Getenv("POD_NAME"))
	return ctrl.Result{}, nil
}
