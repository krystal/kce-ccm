package kce

import (
	"io"

	cloudprovider "k8s.io/cloud-provider"
)

const (
	ProviderName = "kce"
)

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		return newCloudProviderInterface(config)
	})
}
