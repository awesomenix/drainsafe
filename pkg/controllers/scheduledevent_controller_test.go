// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers_test

import (
	"context"
	"testing"
	"time"

	"github.com/awesomenix/drainsafe/pkg/annotations"
	"github.com/awesomenix/drainsafe/pkg/azure"
	"github.com/awesomenix/drainsafe/pkg/controllers"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	scheduledevent = `{
		"DocumentIncarnation": 1,
		"Events": [
			{
			"EventId": "F3E6E2D2-E86A-47F0-AA8E-18918049A2B1",
			"EventStatus": "Scheduled",
			"EventType": "Reboot",
			"ResourceType": "VirtualMachine",
			"Resources": [
				"controlplane_0"
			],
			"NotBefore": "Sun, 30 Jun 2019 16:22:03 GMT"
			}
		]
	}`
)

type testQuery struct {
	getErr  error
	postErr error
	get     string
}

func (q *testQuery) Post(url string, body []byte) error {
	return q.postErr
}

func (q *testQuery) Get(url string) (string, error) {
	return q.get, q.getErr
}

func TestProcessScheduledEvent(t *testing.T) {
	assert := assert.New(t)
	f := fake.NewFakeClient()
	corev1.AddToScheme(scheme.Scheme)
	tQuery := &testQuery{get: scheduledevent}
	c := azure.NewWithQuery(tQuery)

	reconciler := &controllers.ScheduledEventReconciler{
		Client:         f,
		Recorder:       &record.FakeRecorder{},
		Log:            ctrl.Log,
		AzClient:       c,
		Hostname:       "dummyhostname",
		VMInstanceName: "controlplane_0",
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "dummyhostname",
			Annotations: make(map[string]string),
		},
	}
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Running
	err := f.Create(context.TODO(), node)
	assert.Nil(err)
	err = reconciler.ProcessScheduledEvent()
	assert.Nil(err)
	err = f.Get(context.TODO(), types.NamespacedName{Name: node.Name}, node)
	assert.Nil(err)
	assert.Equal("Reboot", node.Annotations[annotations.DrainSafeMaintenanceType])
	assert.Equal(annotations.Scheduled, node.Annotations[annotations.DrainSafeMaintenance])
}

func TestProcessNodeEvent(t *testing.T) {
	assert := assert.New(t)
	f := fake.NewFakeClient()
	corev1.AddToScheme(scheme.Scheme)
	tQuery := &testQuery{get: scheduledevent}
	c := azure.NewWithQuery(tQuery)

	reconciler := &controllers.ScheduledEventReconciler{
		Client:         f,
		Recorder:       &record.FakeRecorder{},
		Log:            ctrl.Log,
		AzClient:       c,
		Hostname:       "dummyhostname",
		VMInstanceName: "controlplane_0",
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "dummyhostname",
			Annotations: make(map[string]string),
		},
	}
	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Running
	err := f.Create(context.TODO(), node)
	assert.Nil(err)
	res, err := reconciler.ProcessNodeEvent(node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{RequeueAfter: 30 * time.Second})

	node.Annotations[annotations.DrainSafeMaintenance] = annotations.Drained
	err = f.Update(context.TODO(), node)
	assert.Nil(err)

	res, err = reconciler.ProcessNodeEvent(node)
	assert.Nil(err)
	assert.Equal(res, ctrl.Result{})
	err = f.Get(context.TODO(), types.NamespacedName{Name: node.Name}, node)
	assert.Nil(err)
	assert.Equal(annotations.Started, node.Annotations[annotations.DrainSafeMaintenance])
}
