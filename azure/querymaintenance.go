// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package azure

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger = ctrl.Log.WithName("azure")

var _ Query = &query{}

// Query interface
type Query interface {
	Post(url string, body []byte) error
	Get(url string) (string, error)
}

type query struct{}

// Client query maintenance client
type Client struct {
	q Query
}

// New create query maintenance client
func New() *Client {
	return &Client{
		q: &query{},
	}
}

// NewWithQuery create query maintenance client with query override
func NewWithQuery(q Query) *Client {
	return &Client{
		q: q,
	}
}

// Post url with body
func (c *query) Post(url string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header = http.Header{
		"Metadata": {"true"},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err, "failed to vm instance name")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode > 299 {
		return errors.Errorf("received non success error code %d", resp.StatusCode)
	}
	return nil
}

// Get url
func (c *query) Get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header = http.Header{
		"Metadata": {"true"},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err, "failed to vm instance name")
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 ||
		resp.StatusCode > 299 {
		return "", errors.Errorf("received non success error code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "failed to read body")
		return "", err
	}

	return string(body), nil
}

// GetVMInstanceName gets current vmss/availability set instance name
func (c *Client) GetVMInstanceName() (string, error) {
	// curl -H Metadata:true "http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text"
	return c.q.Get("http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text")
}

// {
//   "DocumentIncarnation": 1,
//   "Events": [
//     {
//       "EventId": "F3E6E2D2-E86A-47F0-AA8E-18918049A2B1",
//       "EventStatus": "Scheduled",
//       "EventType": "Reboot",
//       "ResourceType": "VirtualMachine",
//       "Resources": [
//         "controlplane_0"
//       ],
//       "NotBefore": "Sun, 30 Jun 2019 16:22:03 GMT"
//     }
//   ]
// }

// ScheduledEvent signifies each scheduled event
type ScheduledEvent struct {
	EventId      string   `json:"EventId"`
	EventStatus  string   `json:"EventStatus"`
	EventType    string   `json:"EventType"`
	ResourceType string   `json:"ResourceType"`
	Resources    []string `json:"Resources"`
	NotBefore    string   `json:"NotBefore"`
}

// ScheduledEventList list of scheduled events
type ScheduledEventList struct {
	DocumentIncarnation int              `json:"DocumentIncarnation"`
	Events              []ScheduledEvent `json:"Events"`
}

func (c *Client) getScheduledEventList() (*ScheduledEventList, error) {
	// curl -H Metadata:true "http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01"
	body, err := c.q.Get("http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01")
	if err != nil {
		log.Error(err, "failed to get scheduled events")
		return nil, err
	}
	result := &ScheduledEventList{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Error(err, "failed to unmarshal body")
		return nil, err
	}

	return result, nil
}

// IsScheduledEvent check if event is scheduled and returns the scheduled event, else nil
func (c *Client) IsScheduledEvent(vmInstanceName string) (string, error) {
	result, err := c.getScheduledEventList()
	if err != nil {
		return "", err
	}

	for _, event := range result.Events {
		if isScheduled(&event) &&
			isDisruptive(&event) &&
			isVMScheduled(&event, vmInstanceName) {
			return event.EventType, nil
		}
	}

	return "", nil
}

// ApproveScheduledEvent approves scheduled event
func (c *Client) ApproveScheduledEvent(vmInstanceName string) error {
	result, err := c.getScheduledEventList()
	if err != nil {
		return err
	}

	for _, event := range result.Events {
		if isScheduled(&event) &&
			isDisruptive(&event) &&
			isVMScheduled(&event, vmInstanceName) {
			return c.approveEvent(&event)
		}
	}

	return nil
}

func (c *Client) approveEvent(event *ScheduledEvent) error {
	// curl -H Metadata:true -X POST -d '{"StartRequests": [{"EventId": "F3E6E2D2-E86A-47F0-AA8E-18918049A2B1"}]}' http://169.254.169.254/metadata/scheduledevents?api-version=2017-11-01
	message := map[string]interface{}{
		"StartRequests": []map[string]string{
			{
				"EventId": event.EventId,
			},
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Error(err, "failed to marshal message")
		return err
	}

	return c.q.Post("http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01", body)
}

func isDisruptive(event *ScheduledEvent) bool {
	if strings.EqualFold(event.EventType, "Reboot") ||
		strings.EqualFold(event.EventType, "Redeploy") ||
		strings.EqualFold(event.EventType, "Preempt") ||
		strings.EqualFold(event.EventType, "Terminate") {
		return true
	}
	return false
}

func isScheduled(event *ScheduledEvent) bool {
	return event.EventStatus == "Scheduled"
}

func isVMScheduled(event *ScheduledEvent, vmInstanceName string) bool {
	if event.ResourceType != "VirtualMachine" {
		return false
	}
	for _, affectedVM := range event.Resources {
		if strings.EqualFold(affectedVM, vmInstanceName) {
			return true
		}
	}
	return false
}
