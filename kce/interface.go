package kce

import (
	"context"
	"fmt"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
	"github.com/sethvargo/go-envconfig"
	"io"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	"net/url"
)

type Config struct {
	APIKey  string `env:"KATAPULT_API_TOKEN"`
	APIHost string `env:"KATAPULT_API_HOST"`
}

// providerFactory creates any dependencies needed by the provider and passes
// them into New. For now, we will source config from the environment, but we
// k8s CCM provides us with an io.Reader which can be used to read a config
// file.
func providerFactory(_ io.Reader) (cloudprovider.Interface, error) {
	c := Config{}

	err := envconfig.Process(context.TODO(), &c)
	if err != nil {
		return nil, err
	}

	if c.APIKey == "" {
		return nil, fmt.Errorf("api key is not configured")
	}

	apiUrl := katapult.DefaultURL
	if c.APIHost != "" {
		klog.Info("default API base URL overrided", "url", c.APIHost)
		apiUrl, err = url.Parse(c.APIHost)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided api url: %w", err)
		}
	}

	client, err := katapult.New(
		katapult.WithAPIKey(c.APIKey),
		katapult.WithBaseURL(apiUrl),
		katapult.WithUserAgent("kce-ccm"), // TODO: Add version.
	)
	if err != nil {
		return nil, err
	}

	return New(c, core.New(client))
}

// New returns a new provider using the provided dependencies.
func New(c Config, client *core.Client) (cloudprovider.Interface, error) {
	// TODO: Interface for client rather than concrete type <3
	return &provider{
		katapult:     client,
		config:       c,
		loadBalancer: &LoadBalancer{},
	}, nil
}

type provider struct {
	katapult     *core.Client
	config       Config
	loadBalancer *LoadBalancer
}

func (p *provider) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	// TODO: Assess if we actually need anything here
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
