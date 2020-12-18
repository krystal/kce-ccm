package kce

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
)

// LoadBalancer is an abstract, pluggable interface for load balancers.
//
// Cloud provider may chose to implement the logic for
// constructing/destroying specific kinds of load balancers in a
// controller separate from the ServiceController.  If this is the case,
// then {Ensure,Update}LoadBalancer must return the ImplementedElsewhere error.
// For the given LB service, the GetLoadBalancer must return "exists=True" if
// there exists a LoadBalancer instance created by ServiceController.
// In all other cases, GetLoadBalancer must return a NotFound error.
// EnsureLoadBalancerDeleted must not return ImplementedElsewhere to ensure
// proper teardown of resources that were allocated by the ServiceController.
// This can happen if a user changes the type of LB via an update to the resource
// or when migrating from ServiceController to alternate implementation.
// The finalizer on the service will be added and removed by ServiceController
// irrespective of the ImplementedElsewhere error. Additional finalizers for
// LB services must be managed in the alternate implementation.

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{}
}

type LoadBalancer struct {
}

func (lb *LoadBalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	return nil, false, nil
}

func (lb *LoadBalancer) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
	return ""
}

func (lb *LoadBalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	return nil, errors.New("EnsureLoadBalancer not implemented")
}

func (lb *LoadBalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	return errors.New("UpdateLoadBalancer not implemented")
}

func (lb *LoadBalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	return errors.New("EnsureLoadBalancerDeleted not implemented")
}
