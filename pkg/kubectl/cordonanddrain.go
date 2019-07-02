// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package kubectl

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubectldrain "k8s.io/kubernetes/pkg/kubectl/cmd/drain"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var log logr.Logger = ctrl.Log.WithName("kubectl")

var _ Client = &client{}

// Client interface for kubernetes
type Client interface {
	Cordon(vmName string) error
	Drain(vmName string, gracePeriod int) error
	Uncordon(vmName string) error
}

type client struct {
	f       cmdutil.Factory
	streams genericclioptions.IOStreams
}

// New creates a new client
func New() (Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to set up client config")
	}

	streams := genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	return &client{
		f:       cmdutil.NewFactory(&RESTConfigClientGetter{Config: cfg}),
		streams: streams,
	}, nil
}

// Cordon cordons  vmname from kubernetes
func (c *client) Cordon(vmName string) error {
	cordon := kubectldrain.NewCmdCordon(c.f, c.streams)
	options := kubectldrain.NewDrainCmdOptions(c.f, c.streams)
	if err := options.Complete(c.f, cordon, []string{vmName}); err != nil {
		return errors.Wrapf(err, "error setting up cordon")
	}

	log.Info("Cordon", "VMName", vmName)
	if err := options.RunCordonOrUncordon(true); err != nil {
		return errors.Wrapf(err, "error cordoning node")
	}

	return nil
}

// Drain drains vmname from kubernetes
func (c *client) Drain(vmName string, gracePeriod int) error {
	drain := kubectldrain.NewCmdDrain(c.f, c.streams)
	drain.SetArgs([]string{vmName, "--ignore-daemonsets", "--force", "--delete-local-data", fmt.Sprintf("--grace-period=%d", gracePeriod)})

	log.Info("Draining", "VMName", vmName)
	if err := drain.Execute(); err != nil {
		return errors.Wrapf(err, "error draining node")
	}

	return nil
}

// Uncordon uncordons vmname from kubernetes
func (c *client) Uncordon(vmName string) error {
	cordon := kubectldrain.NewCmdCordon(c.f, c.streams)
	options := kubectldrain.NewDrainCmdOptions(c.f, c.streams)
	if err := options.Complete(c.f, cordon, []string{vmName}); err != nil {
		return errors.Wrapf(err, "error setting up cordon")
	}

	log.Info("Uncordon", "VMName", vmName)
	if err := options.RunCordonOrUncordon(false); err != nil {
		return errors.Wrapf(err, "error cordoning node")
	}

	return nil
}
