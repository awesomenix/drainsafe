// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"os"
	"time"

	"github.com/awesomenix/drainsafe/pkg/annotations"
	"github.com/awesomenix/drainsafe/pkg/kubectl"
	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	repairmanclient "github.com/awesomenix/repairman/pkg/client"
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

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancerequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/eviction,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile consumes event
func (r *DrainSafeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("node", req.NamespacedName)

	node := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "faled to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	c, err := kubectl.New()
	if err != nil {
		log.Error(err, "failed to create new kubectl client")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// check if repairman is enabled,
	rclient, err := repairmanclient.New(os.Getenv("POD_NAMESPACE"), r.Client)
	if err != nil {
		log.Error(err, "failed to create repairman client")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}
	isEnabled, err := rclient.IsEnabled("node")
	if err != nil {
		log.Error(err, "failed to check if repairman is enabled")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}
	if !isEnabled {
		rclient = nil
	}

	return r.ProcessNodeEvent(c, rclient, node)
}

// SetupWithManager called from maanger to register reconciler
func (r *DrainSafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

func (r *DrainSafeReconciler) updateNodeState(node *corev1.Node, state string) (ctrl.Result, error) {
	log := r.Log.WithValues("node", node.Name)
	if node.Annotations[annotations.DrainSafeMaintenance] == state {
		return ctrl.Result{}, nil
	}
	node.Annotations[annotations.DrainSafeMaintenance] = state
	if err := r.Update(context.TODO(), node); err != nil {
		log.Error(err, "failed to update node")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.Recorder.Eventf(node, "Normal", state, "%s by %s on %s", node.Name, os.Getenv("POD_NAME"), os.Getenv("NODE_NAME"))
	return ctrl.Result{}, nil
}

func (r *DrainSafeReconciler) getMaintenanceApproval(log logr.Logger, rclient *repairmanclient.Client, node *corev1.Node) (ctrl.Result, error) {
	if rclient == nil {
		return r.updateNodeState(node, annotations.MaintenanceApproved)
	}
	log.Info("maintenance approval", "Name", node.Name)
	isApproved, err := rclient.IsMaintenanceApproved(context.TODO(), node.Name, "node")
	if err != nil {
		log.Error(err, "failed to get maintenance approval from repairman")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}
	if isApproved {
		err = rclient.UpdateMaintenanceState(context.TODO(), node.Name, "node", repairmanv1.InProgress)
		if err != nil {
			log.Error(err, "failed to mark maintenance in progress in repairman")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		return r.updateNodeState(node, annotations.MaintenanceApproved)
	}
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

// ProcessNodeEvent processes node event
func (r *DrainSafeReconciler) ProcessNodeEvent(c kubectl.Client, rclient *repairmanclient.Client, node *corev1.Node) (ctrl.Result, error) {
	if node.Annotations == nil {
		return ctrl.Result{}, nil
	}

	log := r.Log.WithValues("node", node.Name)
	maintenance := node.Annotations[annotations.DrainSafeMaintenance]

	log.Info("got node event",
		"Name", node.Name,
		"Maintenance", maintenance,
		"Annotations", node.Annotations)

	if maintenance == annotations.Scheduled {
		return r.getMaintenanceApproval(log, rclient, node)
	}

	if maintenance == annotations.MaintenanceApproved {
		return r.updateNodeState(node, annotations.Cordoning)
	}

	if maintenance == annotations.Cordoning {
		if !node.Spec.Unschedulable {
			if err := c.Cordon(node.Name); err != nil {
				log.Error(err, "failed to cordon vm")
				return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
			}
		}
		return r.updateNodeState(node, annotations.Cordoned)
	}

	if maintenance == annotations.Cordoned {
		return r.updateNodeState(node, annotations.Draining)
	}

	if maintenance == annotations.Draining {
		maintenanceType := node.Annotations[annotations.DrainSafeMaintenanceType]
		if err := c.Drain(node.Name, getGraceTimeoutPeriod(maintenanceType)); err != nil {
			log.Error(err, "failed to drain vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		return r.updateNodeState(node, annotations.Drained)
	}

	if maintenance == annotations.Running {
		if !node.Spec.Unschedulable {
			return ctrl.Result{}, nil
		}
		if rclient != nil {
			if err := rclient.UpdateMaintenanceState(context.TODO(), node.Name, "node", repairmanv1.Completed); err != nil {
				log.Error(err, "failed to mark maintenance in progress in repairman")
				return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
			}
		}
		if err := c.Uncordon(node.Name); err != nil {
			log.Error(err, "failed to cordon vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(node, "Normal", annotations.Uncordoned, "%s by %s on %s", node.Name, os.Getenv("POD_NAME"), os.Getenv("NODE_NAME"))
	}

	return ctrl.Result{}, nil
}

func getGraceTimeoutPeriod(maintenanceType string) int {
	switch maintenanceType {
	case "Reboot", "Freeze":
		return 840
	case "Redeploy":
		return 540
	case "Preempt":
		return 15
	}
	return 60
}
