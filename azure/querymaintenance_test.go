// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package azure_test

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/awesomenix/drainsafe/azure"
	"github.com/stretchr/testify/assert"
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

func TestGetVMInstanceName(t *testing.T) {
	assert := assert.New(t)
	tQuery := &testQuery{get: "dummyvmname"}
	c := azure.NewWithQuery(tQuery)

	vmName, err := c.GetVMInstanceName()
	assert.Equal(vmName, "dummyvmname")
	assert.Nil(err)
	tQuery.getErr = errors.New("dummyerror")
	_, err = c.GetVMInstanceName()
	assert.NotNil(err)
}

func TestIsScheduledEvent(t *testing.T) {
	assert := assert.New(t)
	tQuery := &testQuery{get: scheduledevent}
	c := azure.NewWithQuery(tQuery)

	isScheduled, err := c.IsScheduledEvent("dummyinstancename")
	assert.Empty(isScheduled)
	assert.Nil(err)
	isScheduled, err = c.IsScheduledEvent("controlplane_0")
	assert.Equal(isScheduled, "Reboot")
	assert.Nil(err)

	tQuery.get = "{malformedurl"
	_, err = c.IsScheduledEvent("controlplane_0")
	assert.NotNil(err)
}

func TestApproveScheduledEvent(t *testing.T) {
	assert := assert.New(t)
	tQuery := &testQuery{get: scheduledevent, postErr: errors.New("dummy")}
	c := azure.NewWithQuery(tQuery)

	err := c.ApproveScheduledEvent("dummyinstancename")
	assert.Nil(err)

	err = c.ApproveScheduledEvent("controlplane_0")
	assert.NotNil(err)
}
