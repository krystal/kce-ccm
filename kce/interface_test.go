package kce

import (
	"github.com/krystal/go-katapult/core"
	"github.com/sethvargo/go-envconfig"
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

func Test_loadConfig(t *testing.T) {
	tests := []struct {
		name     string
		lookuper envconfig.Lookuper
		wantErr  string
		want     *Config
	}{
		{
			name: "success",
			lookuper: envconfig.MapLookuper(map[string]string{
				"KATAPULT_API_TOKEN":        "atoken",
				"KATAPULT_API_HOST":         "api.katapult.org",
				"KATAPULT_ORGANIZATION_RID": "fake-org",
				"KATAPULT_DATA_CENTER_RID":  "atlantis",
			}),
			want: &Config{
				APIHost:        "api.katapult.org",
				APIKey:         "atoken",
				OrganizationID: "fake-org",
				DataCenterID:   "atlantis",
			},
		},
		{
			name:     "underlying err propagates",
			lookuper: nil,
			wantErr:  "lookuper cannot be nil",
		},
		{
			name: "api key missing causes error",
			lookuper: envconfig.MapLookuper(map[string]string{
				"KATAPULT_ORGANIZATION_RID": "fake-org",
				"KATAPULT_DATA_CENTER_RID":  "atlantis",
			}),
			wantErr: "api key is not configured",
		},
		{
			name: "org ID missing causes error",
			lookuper: envconfig.MapLookuper(map[string]string{
				"KATAPULT_API_TOKEN":       "atoken",
				"KATAPULT_DATA_CENTER_RID": "atlantis",
			}),
			wantErr: "organization id is not set",
		},
		{
			name: "dc ID missing causes error",
			lookuper: envconfig.MapLookuper(map[string]string{
				"KATAPULT_API_TOKEN":        "atoken",
				"KATAPULT_ORGANIZATION_RID": "fake-org",
			}),
			wantErr: "data center id is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := loadConfig(tt.lookuper)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, c)
		})
	}
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
