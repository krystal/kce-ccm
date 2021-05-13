package kce

import (
	"github.com/krystal/go-katapult/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_orgRef(t *testing.T) {
	c := Config{
		OrganizationID: "a-fake-org",
	}

	assert.Equal(t, &core.Organization{ID: "a-fake-org"}, c.orgRef())
}

func TestConfig_dcRef(t *testing.T) {
	c := Config{
		DataCenterID: "atlantis-central-1",
	}

	assert.Equal(t, &core.DataCenter{ID: "atlantis-central-1"}, c.dcRef())
}

func TestProvider_LoadBalancer(t *testing.T) {
	lbm := &loadBalancerManager{}
	p := &provider{loadBalancer: lbm}

	gotLbm, isSupported := p.LoadBalancer()
	assert.Equal(t, lbm, gotLbm)
	assert.True(t, isSupported)
}

func TestProvider_Instances(t *testing.T) {
	p := &provider{}

	instances, isSupported := p.Instances()
	assert.Nil(t, instances)
	assert.False(t, isSupported)
}

func TestProvider_InstancesV2(t *testing.T) {
	p := &provider{}

	instances, isSupported := p.InstancesV2()
	assert.Nil(t, instances)
	assert.False(t, isSupported)
}

func TestProvider_Zones(t *testing.T) {
	p := &provider{}

	zones, isSupported := p.Zones()
	assert.Nil(t, zones)
	assert.False(t, isSupported)
}

func TestProvider_Clusters(t *testing.T) {
	p := &provider{}

	clusters, isSupported := p.Clusters()
	assert.Nil(t, clusters)
	assert.False(t, isSupported)
}

func TestProvider_Routes(t *testing.T) {
	p := &provider{}

	routes, isSupported := p.Routes()
	assert.Nil(t, routes)
	assert.False(t, isSupported)
}

func TestProvider_ProviderName(t *testing.T) {
	p := &provider{}
	assert.Equal(t, ProviderName, p.ProviderName())
}

func TestProvider_HasClusterID(t *testing.T) {
	p := &provider{}
	assert.True(t, p.HasClusterID())
}
