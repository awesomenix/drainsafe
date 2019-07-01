// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"os"
	"time"

	"github.com/awesomenix/drainsafe/pkg/annotations"
	"github.com/awesomenix/drainsafe/pkg/kubectl"
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

	return r.ProcessNodeEvent(c, node)
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

// ProcessNodeEvent processes node event
func (r *DrainSafeReconciler) ProcessNodeEvent(c kubectl.Client, node *corev1.Node) (ctrl.Result, error) {
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
		if err := c.Drain(node.Name); err != nil {
			log.Error(err, "failed to drain vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		return r.updateNodeState(node, annotations.Drained)
	}

	if maintenance == annotations.Running {
		if !node.Spec.Unschedulable {
			return ctrl.Result{}, nil
		}
		if err := c.Uncordon(node.Name); err != nil {
			log.Error(err, "failed to cordon vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(node, "Normal", annotations.Uncordoned, "%s by %s on %s", node.Name, os.Getenv("POD_NAME"), os.Getenv("NODE_NAME"))
	}

	return ctrl.Result{}, nil
}
