package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func getVMInstanceName() (string, error) {
	// curl -H Metadata:true "http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text"
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text", nil)
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

type resourceEvent struct {
	EventId      string   `json:"EventId"`
	EventStatus  string   `json:"EventStatus"`
	EventType    string   `json:"EventType"`
	ResourceType string   `json:"ResourceType"`
	Resources    []string `json:"Resources"`
	NotBefore    string   `json:"NotBefore"`
}

type scheduledEvent struct {
	DocumentIncarnation int             `json:"DocumentIncarnation"`
	Events              []resourceEvent `json:"Events"`
}

func getScheduledEvent() (*scheduledEvent, error) {
	// curl -H Metadata:true "http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01"
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01", nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{
		"Metadata": {"true"},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err, "failed to vm instance name")
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 ||
		resp.StatusCode > 299 {
		return nil, errors.Errorf("received non success error code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "failed to read body")
		return nil, err
	}
	result := &scheduledEvent{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Error(err, "failed to unmarshal body")
		return nil, err
	}

	return result, nil
}

func isScheduledEvent(vmInstanceName string) (bool, error) {
	result, err := getScheduledEvent()
	if err != nil {
		return false, err
	}

	for _, event := range result.Events {
		if isScheduled(&event) &&
			isDisruptive(&event) &&
			isVMScheduled(&event, vmInstanceName) {
			return true, nil
		}
	}

	return false, nil
}

func approveScheduledEvent(vmInstanceName string) error {
	result, err := getScheduledEvent()
	if err != nil {
		return err
	}

	for _, event := range result.Events {
		if isScheduled(&event) &&
			isDisruptive(&event) &&
			isVMScheduled(&event, vmInstanceName) {
			return approveEvent(&event)
		}
	}

	return nil
}

func approveEvent(event *resourceEvent) error {
	// curl -H Metadata:true -X POST -d '{"StartRequests": [{"EventId": "F3E6E2D2-E86A-47F0-AA8E-18918049A2B1"}]}' http://169.254.169.254/metadata/scheduledevents?api-version=2017-11-01
	message := map[string]interface{}{
		"StartRequests": []map[string]string{
			map[string]string{
				"EventId": event.EventId,
			},
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Error(err, "failed to marshal message")
		return err
	}
	req, err := http.NewRequest("POST", "http://169.254.169.254/metadata/scheduledevents?api-version=2017-08-01", bytes.NewBuffer(body))
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

func isDisruptive(event *resourceEvent) bool {
	return event.EventType == "Reboot" ||
		event.EventType == "Redeploy" ||
		event.EventType == "Preempt"
}

func isScheduled(event *resourceEvent) bool {
	return event.EventStatus == "Scheduled"
}

func isVMScheduled(event *resourceEvent, vmInstanceName string) bool {
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
