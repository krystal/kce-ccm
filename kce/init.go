package kce

import (
	"io"

	cloudprovider "k8s.io/cloud-provider"
)

const (
	ProviderName = "kce"
)

func init() {
	// TODO: Evaluate possibility of removing registry here, and directly referring to this cloud provider interface.
	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		return newCloudProviderInterface(config)
	})
}
