package kce

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
	v1 "k8s.io/api/core/v1"
)

// loadBalancerManager is an abstract, pluggable interface for load balancers.
//
// Cloud provider may chose to implement the logic for
// constructing/destroying specific kinds of load balancers in a
// controller separate from the ServiceController.  If this is the case,
// then {Ensure,Update}loadBalancerManager must return the ImplementedElsewhere error.
// For the given LB service, the GetLoadBalancer must return "exists=True" if
// there exists a loadBalancerManager instance created by ServiceController.
// In all other cases, GetLoadBalancer must return a NotFound error.
// EnsureLoadBalancerDeleted must not return ImplementedElsewhere to ensure
// proper teardown of resources that were allocated by the ServiceController.
// This can happen if a user changes the type of LB via an update to the resource
// or when migrating from ServiceController to alternate implementation.
// The finalizer on the service will be added and removed by ServiceController
// irrespective of the ImplementedElsewhere error. Additional finalizers for
// LB services must be managed in the alternate implementation.

type loadBalancerController interface {
	List(ctx context.Context, org core.OrganizationRef, opts *core.ListOptions) ([]*core.LoadBalancer, *katapult.Response, error)
	Delete(ctx context.Context, lb core.LoadBalancerRef) (*core.LoadBalancer, *katapult.Response, error)
	Update(ctx context.Context, lb core.LoadBalancerRef, args *core.LoadBalancerUpdateArguments) (*core.LoadBalancer, *katapult.Response, error)
	Create(ctx context.Context, org core.OrganizationRef, args *core.LoadBalancerCreateArguments) (*core.LoadBalancer, *katapult.Response, error)
}

type loadBalancerRuleController interface {
	List(ctx context.Context, lb core.LoadBalancerRef, opts *core.ListOptions) ([]core.LoadBalancerRule, *katapult.Response, error)
	Delete(ctx context.Context, lbr core.LoadBalancerRuleRef) (*core.LoadBalancerRule, *katapult.Response, error)
	Update(ctx context.Context, rule core.LoadBalancerRuleRef, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error)
	Create(ctx context.Context, org core.LoadBalancerRef, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error)
}

type loadBalancerManager struct {
	log logr.Logger

	config                     Config
	loadBalancerController     loadBalancerController
	loadBalancerRuleController loadBalancerRuleController
}

var lbNotFound = fmt.Errorf("lb not found")

// listLoadBalancers fetches all LBs for the associated org, paging where necessary
func (lbm *loadBalancerManager) listLoadBalancers(ctx context.Context) ([]*core.LoadBalancer, error) {
	list, resp, err := lbm.loadBalancerController.List(ctx, lbm.config.orgRef(), nil)
	if err != nil {
		return nil, err
	}

	for page := 2; page <= resp.Pagination.TotalPages; page++ {
		more, _, err := lbm.loadBalancerController.List(ctx, lbm.config.orgRef(), &core.ListOptions{Page: page})
		if err != nil {
			return nil, err
		}
		list = append(list, more...)
	}

	return list, err
}

// listLoadBalancerRules fetches all LBRs for an LB, paging where necessary
func (lbm *loadBalancerManager) listLoadBalancerRules(ctx context.Context, lb core.LoadBalancerRef) ([]core.LoadBalancerRule, error) {
	list, resp, err := lbm.loadBalancerRuleController.List(ctx, lb, nil)
	if err != nil {
		return nil, err
	}

	for page := 2; page <= resp.Pagination.TotalPages; page++ {
		more, _, err := lbm.loadBalancerRuleController.List(ctx, lb, &core.ListOptions{Page: page})
		if err != nil {
			return nil, err
		}
		list = append(list, more...)
	}

	return list, err
}

// getLoadBalancer lists all load balancers for an organisation and attemots to
// find a specific load balancer. This will eventually be replaced with a
// bespoke API field to avoid this.
func (lbm *loadBalancerManager) getLoadBalancer(ctx context.Context, name string) (*core.LoadBalancer, error) {
	list, err := lbm.listLoadBalancers(ctx)
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
	// we want to produce a deterministic load balancer name from the service
	// katapult has a limit of 60 characters on name length
	ns := ""
	if service.Namespace != "default" {
		ns = fmt.Sprintf("%s-", service.Namespace)
	}
	untrimmed := fmt.Sprintf("k8s-%s-%s%s", clusterName, ns, service.Name)

	const trimLength = 60
	if len(untrimmed) > trimLength {
		return untrimmed[0:trimLength]
	}
	return untrimmed
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lbm *loadBalancerManager) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	foundLb, err := lbm.getLoadBalancer(ctx, loadBalancerName(clusterName, service))
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
func (lbm *loadBalancerManager) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	return loadBalancerName(clusterName, service)
}

// tidyLoadBalancerRules deletes rules that are no longer in use by the service
// TODO: Instrumentation for number of entities cleaned up
func (lbm *loadBalancerManager) tidyLoadBalancerRules(ctx context.Context, service *v1.Service, lb *core.LoadBalancer) error {
	rules, err := lbm.listLoadBalancerRules(ctx, lb.Ref())
	if err != nil {
		return err
	}

	for _, rule := range rules {
		inUse := false
		for _, servicePort := range service.Spec.Ports {
			if rule.ListenPort == int(servicePort.Port) {
				inUse = true
				break
			}
		}

		if !inUse {
			lbm.log.Info("deleting unused lb rule",
				"loadBalancerId", lb.ID,
				"serviceId", service.UID,
				"ruleId", rule.ID,
				"rulePort", rule.ListenPort,
			)
			_, _, err := lbm.loadBalancerRuleController.Delete(ctx, rule.Ref())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ensureLoadBalancerRules creates or update LB rules to match the ports exposed
// by a kubernetes service.
// TODO: Instrumentation for number of entities created etc
func (lbm *loadBalancerManager) ensureLoadBalancerRules(ctx context.Context, service *v1.Service, lb *core.LoadBalancer) error {
	rules, err := lbm.listLoadBalancerRules(ctx, lb.Ref())
	if err != nil {
		return err
	}

	for _, servicePort := range service.Spec.Ports {
		// attempt to match existing rule to service port based on Port and ListenPort
		var foundRule *core.LoadBalancerRule
		for _, rule := range rules {
			if rule.ListenPort == int(servicePort.Port) {
				foundRule = &rule
				break
			}
		}

		proxyProtocol := false
		checkEnabled := true
		lbRuleArgs := core.LoadBalancerRuleArguments{
			Algorithm:       core.RoundRobinRuleAlgorithm,
			DestinationPort: int(servicePort.NodePort),
			ListenPort:      int(servicePort.Port),
			Protocol:        core.TCPProtocol,
			ProxyProtocol:   &proxyProtocol,
			CheckEnabled:    &checkEnabled,
			CheckProtocol:   core.TCPProtocol,
			CheckTimeout:    5,
			CheckInterval:   10,
			CheckRise:       1,
			CheckFall:       1,
		}

		if foundRule == nil {
			lbm.log.Info("creating lb rule",
				"loadBalancerId", lb.ID,
				"serviceId", service.UID,
				"servicePort", servicePort.Port,
				"servicePortName", servicePort.Name,
				"servicePortTarget", servicePort.TargetPort,
			)
			_, _, err := lbm.loadBalancerRuleController.Create(ctx, lb.Ref(), lbRuleArgs)
			if err != nil {
				return err
			}
		} else {
			lbm.log.Info("updating lb rule",
				"loadBalancerId", lb.ID,
				"ruleId", foundRule.ID,
				"serviceId", service.UID,
				"servicePort", servicePort.Port,
				"servicePortName", servicePort.Name,
				"servicePortTarget", servicePort.TargetPort,
			)
			lbm.log.V(4).Info("updating lb rule",
				"loadBalancerId", lb.ID,
				"ruleId", foundRule.ID,
				"args", lbRuleArgs,
			)
			// TODO: Matcher to avoid unnecessary updates
			_, _, err := lbm.loadBalancerRuleController.Update(ctx, foundRule.Ref(), lbRuleArgs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lbm *loadBalancerManager) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	name := loadBalancerName(clusterName, service)
	lb, err := lbm.getLoadBalancer(ctx, name)
	if err != nil && err != lbNotFound {
		return nil, err
	}

	// If load balancer doesn't exist create it
	if lb == nil {
		lbm.log.Info("creating lb",
			"serviceId", service.UID,
		)
		lb, _, err = lbm.loadBalancerController.Create(ctx, lbm.config.orgRef(), &core.LoadBalancerCreateArguments{
			Name:         name,
			DataCenter:   lbm.config.dcRef(),
			ResourceType: core.VirtualMachineGroupsResourceType,
			ResourceIDs:  &[]string{lbm.config.NodeTagID},
		})
		if err != nil {
			return nil, err
		}
	} else {
		lbm.log.Info("found existing lb",
			"serviceId", service.UID,
			"loadBalancerId", lb.ID,
		)
	}
	// If it already exists, there's not many fields we need to update
	// We do need to update the associated loadBalancerManager rules though.

	err = lbm.ensureLoadBalancerRules(ctx, service, lb)
	if err != nil {
		return nil, err
	}

	err = lbm.tidyLoadBalancerRules(ctx, service, lb)
	if err != nil {
		return nil, err
	}

	return &v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{
		{
			IP: lb.IPAddress.Address,
		},
	}}, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lbm *loadBalancerManager) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	// TODO: Evaluate if required. For now rely on health checking and tagging of hosts.
	return nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (lbm *loadBalancerManager) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	name := loadBalancerName(clusterName, service)
	balancer, err := lbm.getLoadBalancer(ctx, name)
	if err != nil {
		if err == lbNotFound { // If it doesn't exist, good!
			return nil
		}
		return err
	}

	_, _, err = lbm.loadBalancerController.Delete(ctx, balancer.Ref())
	return err
}
