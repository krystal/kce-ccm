package kce

import (
	cloudprovider "k8s.io/cloud-provider"
)

const (
	ProviderName = "kce"
)

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, providerFactory)
}
