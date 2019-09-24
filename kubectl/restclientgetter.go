// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package kubectl

import (
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	cacheddisk "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

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
