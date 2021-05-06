package kce

import (
	"io"
	cloudprovider "k8s.io/cloud-provider"
)

type Config struct {
	APIKey string
}

func initProvider(_ io.Reader) (cloudprovider.Interface, error) {
	c := Config{}
	// TODO: Fetch environment variables
	// TODO: Construct go-katapult library with authorisation setup
	return New(c)
}

func New(c Config) (cloudprovider.Interface, error) {
	return &provider{
		config:       c,
		loadBalancer: &LoadBalancer{},
	}, nil
}

type provider struct {
	config       Config
	loadBalancer *LoadBalancer
}

func (p *provider) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

func (p *provider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return p.loadBalancer, true
}

func (p *provider) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

func (p *provider) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return nil, false
}

func (p *provider) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (p *provider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (p *provider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (p *provider) ProviderName() string {
	return ProviderName
}

func (p *provider) HasClusterID() bool {
	return true
}
