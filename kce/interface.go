package kce

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
	"github.com/sethvargo/go-envconfig"
	"io"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2/klogr"
	"net/url"
)

type Config struct {
	APIKey  string `env:"KATAPULT_API_TOKEN"`
	APIHost string `env:"KATAPULT_API_HOST"`

	OrganizationID string `env:"KATAPULT_ORGANIZATION_RID"`
	DataCenterID   string `env:"KATAPULT_DATA_CENTER_RID"`
}

func (c Config) orgRef() *core.Organization {
	return &core.Organization{
		ID: c.OrganizationID,
	}
}

func (c Config) dcRef() *core.DataCenter {
	return &core.DataCenter{
		ID: c.DataCenterID,
	}
}

func loadConfig(lookuper envconfig.Lookuper) (*Config, error) {
	c := Config{}

	err := envconfig.ProcessWith(context.TODO(), &c, lookuper)
	if err != nil {
		return nil, err
	}

	if c.APIKey == "" {
		return nil, fmt.Errorf("api key is not configured")
	}

	if c.OrganizationID == "" {
		return nil, fmt.Errorf("organization id is not set")
	}

	if c.DataCenterID == "" {
		return nil, fmt.Errorf("data center id is not set")
	}

	return &c, nil
}

// providerFactory creates any dependencies needed by the provider and passes
// them into New. For now, we will source config from the environment, but we
// k8s CCM provides us with an io.Reader which can be used to read a config
// file.
func providerFactory(_ io.Reader) (cloudprovider.Interface, error) {
	log := klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	c, err := loadConfig(envconfig.OsLookuper())
	if err != nil {
		return nil, err
	}

	apiUrl := katapult.DefaultURL
	if c.APIHost != "" {
		log.Info("default API base URL overrided",
			"url", c.APIHost)
		apiUrl, err = url.Parse(c.APIHost)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided api url: %w", err)
		}
	}

	rm, err := katapult.New(
		katapult.WithAPIKey(c.APIKey),
		katapult.WithBaseURL(apiUrl),
		katapult.WithUserAgent("kce-ccm"), // TODO: Add version.
	)
	if err != nil {
		return nil, err
	}
	client := core.New(rm)

	return &provider{
		log:      log,
		katapult: client,
		config:   *c,
		loadBalancer: &loadBalancerManager{
			log:                        log,
			config:                     *c,
			loadBalancerController:     client.LoadBalancers,
			loadBalancerRuleController: client.LoadBalancerRules,
		},
	}, nil
}

type provider struct {
	log          logr.Logger
	katapult     *core.Client
	config       Config
	loadBalancer *loadBalancerManager
}

func (p *provider) Initialize(
	clientBuilder cloudprovider.ControllerClientBuilder,
	stop <-chan struct{}) {
	// TODO: Assess if we actually need anything here
}

// LoadBalancer returns our implementation of the loadBalancerManager provider
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
