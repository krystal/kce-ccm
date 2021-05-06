package kce

import (
	"io"

	cloudprovider "k8s.io/cloud-provider"
)

func newCloudProviderInterface(config io.Reader) (cloudprovider.Interface, error) {

	loadBalancer := NewLoadBalancer()

	return &providerInterface{
		loadBalancer: loadBalancer,
	}, nil
}

type providerInterface struct {
	loadBalancer *LoadBalancer
}

func (c *providerInterface) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

func (c *providerInterface) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return c.loadBalancer, true
}

func (c *providerInterface) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

func (c *providerInterface) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return nil, false
}

func (c *providerInterface) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (c *providerInterface) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (c *providerInterface) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (c *providerInterface) ProviderName() string {
	return ProviderName
}

func (c *providerInterface) HasClusterID() bool {
	return true
}
