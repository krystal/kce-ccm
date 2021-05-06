package kce

import (
	"io"
	cloudprovider "k8s.io/cloud-provider"
)

type Config struct {
	APIKey string
}

// providerFactory creates any dependencies needed by the provider and passes
// them into New. For now, we will source config from the environment, but we
// k8s CCM provides us with an io.Reader which can be used to read a config
// file.
func providerFactory(_ io.Reader) (cloudprovider.Interface, error) {
	c := Config{}
	// TODO: Fetch environment variables
	// TODO: Construct go-katapult library with authorisation setup
	return New(c)
}

// New returns a new provider using the provided dependencies.
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

// LoadBalancer returns our implementation of the LoadBalancer provider
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
