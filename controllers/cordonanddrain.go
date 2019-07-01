// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package controllers

import (
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	cacheddisk "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	kubectldrain "k8s.io/kubernetes/pkg/kubectl/cmd/drain"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var log logr.Logger = ctrl.Log.WithName("cordonanddrain")

// RESTConfigClientGetter rest client getter
type RESTConfigClientGetter struct {
	Config *rest.Config
}

// ToRESTConfig to rest config
func (r *RESTConfigClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.Config, nil
}

// ToDiscoveryClient to discovery client
func (r *RESTConfigClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := r.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return cacheddisk.NewCachedDiscoveryClientForConfig(config, os.TempDir(), "", 10*time.Minute)
}

// ToRESTMapper to rest mapper
func (r *RESTConfigClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	client, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(client), nil
}

// ToRawKubeConfigLoader to raw kubeconfig loader
func (r *RESTConfigClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return &clientcmd.DefaultClientConfig
}

func getOptions(vmName string) (*kubectldrain.DrainCmdOptions, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to set up client config")
	}

	streams := genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	f := cmdutil.NewFactory(&RESTConfigClientGetter{Config: cfg})
	drain := kubectldrain.NewCmdDrain(f, streams)
	options := kubectldrain.NewDrainCmdOptions(f, streams)
	err = drain.ParseFlags([]string{"--ignore-daemonsets", "--force", "--delete-local-data", "--grace-period=60"})
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing flags")
	}
	err = options.Complete(f, drain, []string{vmName})
	if err != nil {
		return nil, errors.Wrapf(err, "error setting up drain")
	}

	return options, nil
}

// Cordon cordons  vmname from kubernetes
func Cordon(vmName string) error {
	options, err := getOptions(vmName)
	if err != nil {
		return errors.Wrapf(err, "error getting options")
	}

	log.Info("Cordon", "VMName", vmName)
	err = options.RunCordonOrUncordon(true)
	if err != nil {
		return errors.Wrapf(err, "error cordoning node")
	}

	return nil
}

// Drain drains vmname from kubernetes
func Drain(vmName string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to set up client config")
	}

	streams := genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	f := cmdutil.NewFactory(&RESTConfigClientGetter{Config: cfg})
	drain := kubectldrain.NewCmdDrain(f, streams)
	drain.SetArgs([]string{vmName, "--ignore-daemonsets", "--force", "--delete-local-data", "--grace-period=60"})

	log.Info("Draining", "VMName", vmName)

	err = drain.Execute()
	if err != nil {
		return errors.Wrapf(err, "error draining node")
	}

	return nil
}

// Uncordon uncordons vmname from kubernetes
func Uncordon(vmName string) error {
	options, err := getOptions(vmName)
	if err != nil {
		return errors.Wrapf(err, "error getting options")
	}

	log.Info("Uncordon", "VMName", vmName)
	err = options.RunCordonOrUncordon(false)
	if err != nil {
		return errors.Wrapf(err, "error cordoning node")
	}

	return nil
}
