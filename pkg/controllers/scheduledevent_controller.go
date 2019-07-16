// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/awesomenix/drainsafe/pkg/annotations"
	"github.com/awesomenix/drainsafe/pkg/azure"
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
	StopCh         <-chan struct{}
	AzClient       *azure.Client
	Hostname       string
	VMInstanceName string
}

// Reconcile consumes event
func (r *ScheduledEventReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if !strings.EqualFold(req.Name, r.Hostname) {
		return ctrl.Result{}, nil
	}

	ctx := context.Background()
	log := r.Log.WithValues("node", req.NamespacedName)

	node := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	return r.ProcessNodeEvent(node)
}

// SetupWithManager called from manager to register reconciler
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

	r.AzClient = azure.New()
	vmInstanceName, err := r.AzClient.GetVMInstanceName()
	if err != nil {
		r.Log.Error(err, "failed to get vm instance name")
		return err
	}

	r.Hostname = hostname
	r.VMInstanceName = vmInstanceName

	go r.eventWatcher()

	return nil
}

func (r *ScheduledEventReconciler) eventWatcher() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.ProcessScheduledEvent()
		case <-r.StopCh:
			return
		}
	}
}

func (r *ScheduledEventReconciler) updateNodeState(node *corev1.Node, state string) (ctrl.Result, error) {
	if node.Annotations[annotations.DrainSafeMaintenance] == state {
		return ctrl.Result{}, nil
	}
	r.Log.Info("updating node state", "Current", node.Annotations[annotations.DrainSafeMaintenance], "Desired", state)
	node.Annotations[annotations.DrainSafeMaintenance] = state
	if err := r.Update(context.TODO(), node); err != nil {
		r.Log.Error(err, "failed to update node")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.Recorder.Eventf(node, "Normal", state, "%s by %s", node.Name, os.Getenv("POD_NAME"))
	return ctrl.Result{}, nil
}

func (r *ScheduledEventReconciler) updateNodeStateWithType(node *corev1.Node, state, mtype string) (ctrl.Result, error) {
	if node.Annotations[annotations.DrainSafeMaintenance] == state {
		return ctrl.Result{}, nil
	}
	r.Log.Info("updating node state", "Current", node.Annotations[annotations.DrainSafeMaintenance], "Desired", state, "MaintenanceType", mtype)
	node.Annotations[annotations.DrainSafeMaintenance] = state
	node.Annotations[annotations.DrainSafeMaintenanceType] = mtype
	if err := r.Update(context.TODO(), node); err != nil {
		r.Log.Error(err, "failed to update node")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.Recorder.Eventf(node, "Normal", state, "%s on %s by %s", mtype, node.Name, os.Getenv("POD_NAME"))
	return ctrl.Result{}, nil
}

// ProcessNodeEvent processes node event
func (r *ScheduledEventReconciler) ProcessNodeEvent(node *corev1.Node) (ctrl.Result, error) {
	if node.Annotations == nil {
		return ctrl.Result{}, nil
	}

	log := r.Log.WithValues("node", node.Name)
	maintenance := node.Annotations[annotations.DrainSafeMaintenance]

	log.Info("got update event",
		"Name", node.Name,
		"Maintenance", maintenance)

	if maintenance == annotations.Drained {
		if err := r.AzClient.ApproveScheduledEvent(r.VMInstanceName); err != nil {
			log.Error(err, "failed to approve scheduled event")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		return r.updateNodeState(node, annotations.Started)
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// ProcessScheduledEvent process scheduled event.
func (r *ScheduledEventReconciler) ProcessScheduledEvent() error {
	node := &corev1.Node{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: r.Hostname}, node); err != nil {
		r.Log.Error(err, "failed to get node", "Name", r.Hostname)
		return err
	}
	maintenance := node.Annotations[annotations.DrainSafeMaintenance]
	isScheduled, err := r.AzClient.IsScheduledEvent(r.VMInstanceName)
	if err != nil {
		r.Log.Error(err, "failed to find scheduled events")
		return err
	}
	if len(isScheduled) != 0 {
		if maintenance != annotations.Running {
			r.Log.Info("node is under going maintenance, skipping setting annotation", "Maintenance", maintenance)
			return nil
		}
		_, err = r.updateNodeStateWithType(node, annotations.Scheduled, isScheduled)
		return err
	}
	_, err = r.updateNodeStateWithType(node, annotations.Running, "")
	return err
}
