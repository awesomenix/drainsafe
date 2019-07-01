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
	log := r.Log.WithValues("drainsafe", req.NamespacedName)

	// your logic here
	node := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "faled to get event", "NamespacedName", req.NamespacedName.String())
		return ctrl.Result{}, err
	}

	if node.Annotations == nil {
		return ctrl.Result{}, nil
	}

	log.Info("got update event",
		"Name", node.Name,
		"Annotations", node.Annotations)

	maintenance := node.Annotations[annotations.DrainSafeMaintenance]

	if maintenance == annotations.Scheduled {
		return r.updateNodeState(node, annotations.Cordoning)
	}

	if maintenance == annotations.Cordoning {
		if !node.Spec.Unschedulable {
			err = Cordon(node.Name)
			if err != nil {
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
		err = Drain(node.Name)
		if err != nil {
			log.Error(err, "failed to drain vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		return r.updateNodeState(node, annotations.Drained)
	}

	if maintenance == annotations.Running {
		if !node.Spec.Unschedulable {
			return ctrl.Result{}, nil
		}
		err = Uncordon(node.Name)
		if err != nil {
			log.Error(err, "failed to cordon vm")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		r.Recorder.Eventf(node, "Normal", annotations.Uncordoned, "%s", node.Name)
	}

	return ctrl.Result{}, nil
}

func (r *DrainSafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

func (r *DrainSafeReconciler) updateNodeState(node *corev1.Node, state string) (ctrl.Result, error) {
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
