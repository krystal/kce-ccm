package kce

import (
	"context"
	"errors"
	"fmt"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
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

type loadBalancerController interface {
	List(ctx context.Context, org *core.Organization, opts *core.ListOptions) ([]*core.LoadBalancer, *katapult.Response, error)
	Delete(ctx context.Context, lb *core.LoadBalancer) (*core.LoadBalancer, *katapult.Response, error)
	Update(ctx context.Context, lb *core.LoadBalancer, args *core.LoadBalancerUpdateArguments) (*core.LoadBalancer, *katapult.Response, error)
	Create(ctx context.Context, org *core.Organization, args *core.LoadBalancerCreateArguments) (*core.LoadBalancer, *katapult.Response, error)
}

type loadBalancerRuleController interface {
}

type LoadBalancer struct {
	config                     Config
	loadBalancerController     loadBalancerController
	loadBalancerRuleController loadBalancerRuleController
}

var lbNotFound = fmt.Errorf("lb not found")

// getLoadBalancer lists all load balancers for an organisation and attemots to
// find a specific load balancer. This will eventually be replaced with a
// bespoke API field to avoid this.
func (lb *LoadBalancer) getLoadBalancer(ctx context.Context, name string) (*core.LoadBalancer, error) {
	list, _, err := lb.loadBalancerController.List(ctx, lb.config.orgRef(), &core.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	for _, potentialMatch := range list {
		if potentialMatch.Name == name {
			return potentialMatch, nil
		}
	}

	return nil, lbNotFound
}

func loadBalancerName(clusterName string, service *v1.Service) string {
	// kubernetes uid looks like this "b5216b07-2cb4-4429-8294-23883301a01e"
	// we want to produce a deterministic load balancer name from this.
	// katapult has a limit of 255 characters on name length
	return fmt.Sprintf("k8s-%s-%s", clusterName, service.UID)
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lb *LoadBalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	foundLb, err := lb.getLoadBalancer(ctx, loadBalancerName(clusterName, service))
	if err != nil {
		if err == lbNotFound {
			return nil, false, nil
		}

		return nil, false, err
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: foundLb.IPAddress.Address,
			},
		},
	}, true, nil
}

// GetLoadBalancerName returns the name of the load balancer. Implementations
// must treat the *v1.Service parameter as read-only and not modify it.
func (lb *LoadBalancer) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	return loadBalancerName(clusterName, service)
}

// tidyLoadBalancerRules deletes rules that are no longer in use by the service
func (lb *LoadBalancer) tidyLoadBalancerRules(ctx context.Context, service *v1.Service, balancer *core.LoadBalancer) error {
	// 1. Fetch all rules
	// 2. Try to match rule port to service port
	// 3. No match? Delete. Log message and metric.

	return nil
}

// ensureLoadBalancerRules creates or update LB rules to match the ports exposed
// by a kubernetes service.
func (lb *LoadBalancer) ensureLoadBalancerRules(ctx context.Context, service *v1.Service, balancer *core.LoadBalancer) error {
	// Ensure a rule exists for each port :)
	for _, port := range service.Spec.Ports {
		// If rule exists with external port, ensure match or update
		// If no rule exists, create
	}

	return nil
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lb *LoadBalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	// Get
	name := loadBalancerName(clusterName, service)
	balancer, err := lb.getLoadBalancer(ctx, name)
	if err != nil && err != lbNotFound {
		return nil, err
	}

	if balancer == nil {
		balancer, _, err = lb.loadBalancerController.Create(ctx, lb.config.orgRef(), &core.LoadBalancerCreateArguments{
			Name:       name,
			DataCenter: lb.config.dcRef(),
		})
		if err != nil {
			return nil, err
		}
	}

	err = lb.ensureLoadBalancerRules(ctx, service, balancer)
	if err != nil {
		return nil, err
	}

	err = lb.tidyLoadBalancerRules(ctx, service, balancer)
	if err != nil {
		return nil, err
	}

	return &v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{
		{
			IP: balancer.IPAddress.Address,
		},
	}}, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lb *LoadBalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	// This will probably not do much. We will implement load balancers that target the group/tag of nodes rather than specified nodes.
	return errors.New("UpdateLoadBalancer not implemented")
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lb *LoadBalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	name := loadBalancerName(clusterName, service)
	balancer, err := lb.getLoadBalancer(ctx, name)
	if err != nil {
		if err == lbNotFound { // If it doesn't exist, good!
			return nil
		}
		return err
	}

	_, _, err = lb.loadBalancerController.Delete(ctx, balancer)
	return err
}
