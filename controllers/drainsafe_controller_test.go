// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers_test

import (
	"context"
	"time"

	"testing"

	"github.com/awesomenix/drainsafe/annotations"
	"github.com/awesomenix/drainsafe/controllers"
	"github.com/awesomenix/drainsafe/kubectl"
	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	repairmanclient "github.com/awesomenix/repairman/pkg/client"
	repairmantest "github.com/awesomenix/repairman/pkg/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ kubectl.Client = &fakeKubeClient{}

type fakeKubeClient struct {
	cordonerr   error
	drainerr    error
	uncordonerr error
}

func (f *fakeKubeClient) Cordon(vmName string) error {
	return f.cordonerr
}

func (f *fakeKubeClient) Drain(vmName string, gracePeriod int) error {
	return f.drainerr
}

func (f *fakeKubeClient) Uncordon(vmName string) error {
	return f.uncordonerr
}

func TestNoAnnotations(t *testing.T) {
	assert := assert.New(t)
	corev1.AddToScheme(scheme.Scheme)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	reconciler := &controllers.DrainSafeReconciler{
		Client:   f,
		Recorder: &record.FakeRecorder{},
		Log:      ctrl.Log,
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
		},
	}
	_, err := reconciler.ProcessNodeEvent(&fakeKubeClient{}, nil, node)
	assert.Nil(err)
	assert.Equal(len(node.Annotations), 0)
}

func TestReconcile(t *testing.T) {
	assert := assert.New(t)
	f := fake.NewFakeClient()
	corev1.AddToScheme(scheme.Scheme)
	repairmanv1.AddToScheme(scheme.Scheme)
	reconciler := &controllers.DrainSafeReconciler{
		Client:   f,
		Recorder: &record.FakeRecorder{},
		Log:      ctrl.Log,
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "dummynode",
			Annotations: make(map[string]string),
		},
	}
	err := f.Create(context.TODO(), node)
	assert.Nil(err)
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Scheduled
	for _, state := range []string{
		annotations.MaintenanceApproved,
		annotations.Cordoning,
		annotations.Cordoned,
		annotations.Draining,
		annotations.Drained} {
		res, err := reconciler.ProcessNodeEvent(&fakeKubeClient{}, nil, node)
		assert.Nil(err)
		assert.Equal(res, ctrl.Result{})
		assert.Equal(state, node.Annotations[annotations.DrainSafeMaintenance])
		if state == annotations.Cordoned {
			assert.Equal(node.Annotations[annotations.DrainSafeMaintenanceOwner], annotations.Drainsafe)
		}
	}
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Running
	node.Spec.Unschedulable = true
	res, err := reconciler.ProcessNodeEvent(&fakeKubeClient{uncordonerr: errors.New("error")}, nil, node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{RequeueAfter: 1 * time.Minute})
	res, err = reconciler.ProcessNodeEvent(&fakeKubeClient{}, nil, node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{})
	assert.Equal(node.Annotations[annotations.DrainSafeMaintenanceOwner], "")
}

func TestReconcileWithRepairMan(t *testing.T) {
	assert := assert.New(t)
	f := fake.NewFakeClient()
	corev1.AddToScheme(scheme.Scheme)
	repairmanv1.AddToScheme(scheme.Scheme)

	repairmantest.ReconcileML(f, assert)

	reconciler := &controllers.DrainSafeReconciler{
		Client:   f,
		Recorder: &record.FakeRecorder{},
		Log:      ctrl.Log,
	}

	rclient := &repairmanclient.Client{
		Name:       "fakeName",
		Client:     f,
		NewRequest: repairmantest.NewRequest,
	}

	node := &corev1.Node{}
	err := f.Get(context.TODO(), types.NamespacedName{Name: "dummynode0"}, node)
	assert.Nil(err)
	node.Annotations = make(map[string]string)
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Scheduled
	res, err := reconciler.ProcessNodeEvent(&fakeKubeClient{}, rclient, node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{RequeueAfter: 1 * time.Minute})
	assert.NotEqual(annotations.MaintenanceApproved, node.Annotations[annotations.DrainSafeMaintenance])
	repairmantest.ReconcileMR(f)
	for _, state := range []string{
		annotations.MaintenanceApproved,
		annotations.Cordoning,
		annotations.Cordoned,
		annotations.Draining,
		annotations.Drained} {
		res, err := reconciler.ProcessNodeEvent(&fakeKubeClient{}, rclient, node)
		assert.Nil(err)
		assert.Equal(res, ctrl.Result{})
		assert.Equal(state, node.Annotations[annotations.DrainSafeMaintenance])
	}
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Running
	node.Spec.Unschedulable = true
	res, err = reconciler.ProcessNodeEvent(&fakeKubeClient{uncordonerr: errors.New("error")}, rclient, node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{RequeueAfter: 1 * time.Minute})
}
