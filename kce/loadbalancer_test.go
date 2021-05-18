package kce

import (
	"context"
	"fmt"
	logTest "github.com/go-logr/logr/testing"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"testing"
)

type mockLBController struct {
	createdItems int
	items        []core.LoadBalancer
}

func (lbc *mockLBController) List(_ context.Context, _ *core.Organization, opts *core.ListOptions) ([]*core.LoadBalancer, *katapult.Response, error) {
	perPage := 2
	page := 1
	if opts != nil {
		if opts.PerPage != 0 {
			perPage = opts.PerPage
		}
		if opts.Page != 0 {
			page = opts.Page
		}
	}

	pagedOut := make([]*core.LoadBalancer, 0)
	start := (page - 1) * perPage
	end := page * perPage
	if end > len(lbc.items) {
		end = len(lbc.items)
	}
	for i := start; i < end; i++ {
		copyOfItem := lbc.items[i]
		if copyOfItem.ID == "error" {
			return nil, nil, fmt.Errorf("error from %d", i)
		}
		pagedOut = append(pagedOut, &copyOfItem)
	}

	return pagedOut, &katapult.Response{
		Pagination: &katapult.Pagination{
			CurrentPage: page,
			PerPage:     perPage,
			TotalPages:  int(math.Ceil(float64(len(lbc.items)) / float64(perPage))),
			Total:       len(lbc.items),
		},
	}, nil
}

func (lbc *mockLBController) Delete(_ context.Context, lb *core.LoadBalancer) (*core.LoadBalancer, *katapult.Response, error) {
	for i, item := range lbc.items {
		if item.ID == lb.ID {
			lbc.items = append(lbc.items[:i], lbc.items[i+1:]...)
			return &item, &katapult.Response{}, nil
		}
	}

	return nil, nil, fmt.Errorf("tried to delete non-existent element")
}

func (lbc *mockLBController) Update(ctx context.Context, lb *core.LoadBalancer, args *core.LoadBalancerUpdateArguments) (*core.LoadBalancer, *katapult.Response, error) {
	return nil, nil, fmt.Errorf("unimplemented")
}

func (lbc *mockLBController) Create(_ context.Context, _ *core.Organization, args *core.LoadBalancerCreateArguments) (*core.LoadBalancer, *katapult.Response, error) {
	newItem := core.LoadBalancer{
		ID:           fmt.Sprintf("created-%d", lbc.createdItems),
		IPAddress:    &core.IPAddress{Address: fmt.Sprintf("10.0.0.%d", lbc.createdItems)},
		ResourceType: args.ResourceType,
		Name:         args.Name,
	}
	if args.ResourceIDs != nil {
		newItem.ResourceIDs = *args.ResourceIDs
	}
	lbc.items = append(lbc.items, newItem)
	lbc.createdItems++

	return &newItem, &katapult.Response{}, nil
}

type mockLBRController struct {
	createdItems int
	items        []core.LoadBalancerRule
}

func (lbr *mockLBRController) List(_ context.Context, _ *core.LoadBalancer, opts *core.ListOptions) ([]core.LoadBalancerRule, *katapult.Response, error) {
	perPage := 2
	page := 1
	if opts != nil {
		if opts.PerPage != 0 {
			perPage = opts.PerPage
		}
		if opts.Page != 0 {
			page = opts.Page
		}
	}

	pagedOut := make([]core.LoadBalancerRule, 0)
	start := (page - 1) * perPage
	end := page * perPage
	if end > len(lbr.items) {
		end = len(lbr.items)
	}
	for i := start; i < end; i++ {
		if lbr.items[i].ID == "error" {
			return nil, nil, fmt.Errorf("error from %d", i)
		}
		pagedOut = append(pagedOut, lbr.items[i])
	}

	return pagedOut, &katapult.Response{
		Pagination: &katapult.Pagination{
			CurrentPage: page,
			PerPage:     perPage,
			TotalPages:  int(math.Ceil(float64(len(lbr.items)) / float64(perPage))),
			Total:       len(lbr.items),
		},
	}, nil
}

func (lbrc *mockLBRController) Delete(ctx context.Context, lbr *core.LoadBalancerRule) (*core.LoadBalancerRule, *katapult.Response, error) {
	for i, item := range lbrc.items {
		if item.ID == lbr.ID {
			lbrc.items = append(lbrc.items[:i], lbrc.items[i+1:]...)
			return &item, &katapult.Response{}, nil
		}
	}

	return nil, nil, fmt.Errorf("tried to delete non-existent element")
}

// mergeString takes the first non zero value string
func mergeString(a, b string) string {
	if a != "" {
		return a
	}

	return b
}

// mergeInt returns the first non zero int
func mergeInt(a, b int) int {
	if a != 0 {
		return a
	}

	return b
}

func (lbrc *mockLBRController) Update(ctx context.Context, rule *core.LoadBalancerRule, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error) {
	updateItem := core.LoadBalancerRule{}
	updateIndex := -1
	for i, item := range lbrc.items {
		if item.ID == rule.ID {
			updateIndex = i
			updateItem = item
		}
	}
	if updateIndex == -1 {
		return nil, nil, fmt.Errorf("non-existent")
	}

	// merge objects
	updateItem.Algorithm = core.LoadBalancerRuleAlgorithm(mergeString(string(args.Algorithm), string(updateItem.Algorithm)))
	updateItem.DestinationPort = mergeInt(args.DestinationPort, updateItem.DestinationPort)
	updateItem.ListenPort = mergeInt(args.ListenPort, updateItem.ListenPort)
	updateItem.Protocol = core.Protocol(mergeString(string(args.Protocol), string(updateItem.Protocol)))
	if args.ProxyProtocol != nil {
		updateItem.ProxyProtocol = *args.ProxyProtocol
	}
	if args.CheckEnabled != nil {
		updateItem.CheckEnabled = *args.CheckEnabled
	}
	updateItem.CheckFall = mergeInt(args.CheckFall, updateItem.CheckFall)
	updateItem.CheckInterval = mergeInt(args.CheckInterval, updateItem.CheckInterval)
	updateItem.CheckPath = mergeString(args.CheckPath, updateItem.CheckPath)
	updateItem.CheckProtocol = core.Protocol(mergeString(string(args.CheckProtocol), string(updateItem.CheckProtocol)))
	updateItem.CheckRise = mergeInt(args.CheckRise, updateItem.CheckRise)
	updateItem.CheckTimeout = mergeInt(args.CheckTimeout, updateItem.CheckTimeout)
	// set
	lbrc.items[updateIndex] = updateItem

	return &updateItem, &katapult.Response{}, nil
}

func (lbrc *mockLBRController) Create(ctx context.Context, lb *core.LoadBalancer, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error) {
	for _, item := range lbrc.items {
		if item.ListenPort == args.ListenPort {
			return &item, &katapult.Response{}, fmt.Errorf("already existing listen port")
		}
	}

	proxyProtocol := false
	if args.ProxyProtocol != nil {
		proxyProtocol = *args.ProxyProtocol
	}
	checkEnabled := false
	if args.CheckEnabled != nil {
		checkEnabled = *args.CheckEnabled
	}
	newItem := core.LoadBalancerRule{
		ID:              fmt.Sprintf("created-%d", lbrc.createdItems),
		Algorithm:       args.Algorithm,
		DestinationPort: args.DestinationPort,
		ListenPort:      args.ListenPort,
		Protocol:        args.Protocol,
		ProxyProtocol:   proxyProtocol,
		CheckEnabled:    checkEnabled,
		CheckFall:       args.CheckFall,
		CheckInterval:   args.CheckInterval,
		CheckPath:       args.CheckPath,
		CheckProtocol:   args.CheckProtocol,
		CheckRise:       args.CheckRise,
		CheckTimeout:    args.CheckTimeout,
	}
	lbrc.items = append(lbrc.items, newItem)
	lbrc.createdItems++

	return &newItem, &katapult.Response{}, nil
}

func TestLoadBalancerManager_listLoadBalancers(t *testing.T) {
	tests := []struct {
		name string

		seedData []core.LoadBalancer

		want    []*core.LoadBalancer
		wantErr string
	}{
		{
			name: "success",
			seedData: []core.LoadBalancer{
				{ID: "123"},
				{ID: "456"},
				{ID: "789"},
				{ID: "abc"},
				{ID: "def"},
			},
			want: []*core.LoadBalancer{
				{ID: "123"},
				{ID: "456"},
				{ID: "789"},
				{ID: "abc"},
				{ID: "def"},
			},
		},
		{
			name: "immediate error",
			seedData: []core.LoadBalancer{
				{ID: "error"},
			},
			wantErr: "error from 0",
		},
		{
			name: "error in batch",
			seedData: []core.LoadBalancer{
				{ID: "123"},
				{ID: "456"},
				{ID: "error"},
			},
			wantErr: "error from 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBController{items: tt.seedData}
			lbm := loadBalancerManager{
				loadBalancerController: lbc,
				log:                    logTest.TestLogger{T: t},
			}

			got, err := lbm.listLoadBalancers(context.TODO())
			assert.Equal(t, tt.want, got)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_listLoadBalancerRules(t *testing.T) {
	tests := []struct {
		name string

		seedData []core.LoadBalancerRule

		want    []core.LoadBalancerRule
		wantErr string
	}{
		{
			name: "success",
			seedData: []core.LoadBalancerRule{
				{ID: "123"},
				{ID: "456"},
				{ID: "789"},
				{ID: "abc"},
				{ID: "def"},
			},
			want: []core.LoadBalancerRule{
				{ID: "123"},
				{ID: "456"},
				{ID: "789"},
				{ID: "abc"},
				{ID: "def"},
			},
		},
		{
			name: "immediate error",
			seedData: []core.LoadBalancerRule{
				{ID: "error"},
			},
			wantErr: "error from 0",
		},
		{
			name: "error in batch",
			seedData: []core.LoadBalancerRule{
				{ID: "123"},
				{ID: "456"},
				{ID: "error"},
			},
			wantErr: "error from 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbrc := &mockLBRController{items: tt.seedData}
			lbm := loadBalancerManager{
				loadBalancerRuleController: lbrc,
				log:                        logTest.TestLogger{T: t},
			}

			got, err := lbm.listLoadBalancerRules(context.TODO(), &core.LoadBalancer{})
			assert.Equal(t, tt.want, got)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_getLoadBalancer(t *testing.T) {
	tests := []struct {
		name          string
		loadBalancers []core.LoadBalancer

		wantName string

		want    *core.LoadBalancer
		wantErr string
	}{
		{
			name: "success",
			loadBalancers: []core.LoadBalancer{
				{
					Name: "woo",
					ID:   "lb_dkhVsHN8s8OpEeM9",
				},
			},
			wantName: "woo",
			want: &core.LoadBalancer{
				Name: "woo",
				ID:   "lb_dkhVsHN8s8OpEeM9",
			},
		},
		{
			name: "not found",
			loadBalancers: []core.LoadBalancer{
				{
					Name: "woo",
					ID:   "lb_dkhVsHN8s8OpEeM9",
				},
			},
			wantName: "idontexist",
			wantErr:  lbNotFound.Error(),
		},
		{
			name: "err propagates",
			loadBalancers: []core.LoadBalancer{
				{
					ID: "error",
				},
			},
			wantName: "woo",
			wantErr:  "error from 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBController{items: tt.loadBalancers}
			lbm := loadBalancerManager{
				loadBalancerController: lbc,
				log:                    logTest.TestLogger{T: t},
			}

			lb, err := lbm.getLoadBalancer(context.TODO(), tt.wantName)
			assert.Equal(t, tt.want, lb)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_GetLoadBalancer(t *testing.T) {
	tests := []struct {
		name string

		loadBalancers []core.LoadBalancer

		clusterName string
		service     *v1.Service

		wantStatus *v1.LoadBalancerStatus
		wantExists bool
		wantErr    string
	}{
		{
			name: "exists",
			loadBalancers: []core.LoadBalancer{
				{
					Name:      "k8s-test-b5216b07-2cb4-4429-8294-23883301a01e",
					IPAddress: &core.IPAddress{Address: "10.0.0.1"},
				},
			},
			clusterName: "test",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},
			wantStatus: &v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{
				{
					IP: "10.0.0.1",
				},
			}},
			wantExists: true,
		},
		{
			name:        "nonexistent",
			clusterName: "test",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},
			wantExists: false,
		},
		{
			name:        "error",
			clusterName: "test",
			loadBalancers: []core.LoadBalancer{
				{
					ID: "error",
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},
			wantExists: false,
			wantErr:    "error from 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBController{items: tt.loadBalancers}
			lbm := loadBalancerManager{
				loadBalancerController: lbc,
				log:                    logTest.TestLogger{T: t},
			}

			gotStatus, gotExists, gotErr := lbm.GetLoadBalancer(context.TODO(), tt.clusterName, tt.service)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.Equal(t, tt.wantExists, gotExists)
			if tt.wantErr == "" {
				assert.NoError(t, gotErr)
			} else {
				assert.EqualError(t, gotErr, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_GetLoadBalancerName(t *testing.T) {
	lbm := loadBalancerManager{}

	got := lbm.GetLoadBalancerName(
		context.TODO(),
		"big-bad-cluster",
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				UID: "b5216b07-2cb4-4429-8294-23883301a01e",
			},
		},
	)

	want := "k8s-big-bad-cluster-b5216b07-2cb4-4429-8294-23883301a01e"

	assert.Equal(t, want, got)
}

func TestLoadBalancerManager_EnsureLoadBalancerDeleted(t *testing.T) {
	tests := []struct {
		name string

		loadBalancers []core.LoadBalancer

		clusterName string
		service     *v1.Service

		wantLoadBalancers []core.LoadBalancer

		wantErr string
	}{
		{
			name: "deletes",
			loadBalancers: []core.LoadBalancer{
				{
					Name:      "k8s-test-b5216b07-2cb4-4429-8294-23883301a01e",
					IPAddress: &core.IPAddress{Address: "10.0.0.1"},
				},
			},
			clusterName: "test",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},

			wantLoadBalancers: []core.LoadBalancer{},
		},
		{
			name:          "handles non existent",
			loadBalancers: []core.LoadBalancer{},
			clusterName:   "test",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},
			wantLoadBalancers: []core.LoadBalancer{},
		},
		{
			name:        "propagates error",
			clusterName: "test",
			loadBalancers: []core.LoadBalancer{
				{
					ID: "error",
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
			},
			wantLoadBalancers: []core.LoadBalancer{
				{
					ID: "error",
				},
			},
			wantErr: "error from 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBController{items: tt.loadBalancers}
			lbm := loadBalancerManager{
				loadBalancerController: lbc,
				log:                    logTest.TestLogger{T: t},
			}

			err := lbm.EnsureLoadBalancerDeleted(context.TODO(), tt.clusterName, tt.service)
			assert.Equal(t, tt.wantLoadBalancers, lbc.items)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_tidyLoadBalancerRules(t *testing.T) {
	// TODO: Cover error cases
	tests := []struct {
		name string

		loadBalancerRules []core.LoadBalancerRule
		loadBalancer      *core.LoadBalancer

		service *v1.Service

		wantLoadBalancerRules []core.LoadBalancerRule

		wantErr string
	}{
		{
			name: "correctly handles two rules",
			loadBalancerRules: []core.LoadBalancerRule{
				{
					ID:         "willbekept",
					ListenPort: 133,
				},
				{
					ID:         "willbedeleted",
					ListenPort: 1337,
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
				Spec: v1.ServiceSpec{Ports: []v1.ServicePort{
					{
						Port: 133,
					},
				}},
			},
			loadBalancer: &core.LoadBalancer{},

			wantLoadBalancerRules: []core.LoadBalancerRule{
				{
					ID:         "willbekept",
					ListenPort: 133,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBRController{items: tt.loadBalancerRules}
			lbm := loadBalancerManager{
				loadBalancerRuleController: lbc,
				log:                        logTest.TestLogger{T: t},
			}

			err := lbm.tidyLoadBalancerRules(context.TODO(), tt.service, tt.loadBalancer)
			assert.Equal(t, tt.wantLoadBalancerRules, lbc.items)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_ensureLoadBalancerRules(t *testing.T) {
	tests := []struct {
		name string

		loadBalancerRules []core.LoadBalancerRule
		loadBalancer      *core.LoadBalancer

		service *v1.Service

		wantLoadBalancerRules []core.LoadBalancerRule

		wantErr string
	}{
		{
			name: "creates and updates",
			loadBalancerRules: []core.LoadBalancerRule{
				{
					ID:              "lbrule_xICEvzBIgsjyHQQv",
					ListenPort:      144,
					DestinationPort: 132,
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
				Spec: v1.ServiceSpec{Ports: []v1.ServicePort{
					{
						Port:     144,
						NodePort: 1337,
					},
					{
						Port:     256,
						NodePort: 199,
					},
				}},
			},
			loadBalancer: &core.LoadBalancer{
				ID: "lb_npORVDLVrf7MlghA",
			},

			wantLoadBalancerRules: []core.LoadBalancerRule{
				{
					ID:              "lbrule_xICEvzBIgsjyHQQv",
					Algorithm:       core.RoundRobinRuleAlgorithm,
					DestinationPort: 1337,
					ListenPort:      144,
					Protocol:        core.TCPProtocol,
					CheckEnabled:    true,
					CheckFall:       1,
					CheckInterval:   10,
					CheckProtocol:   core.TCPProtocol,
					CheckRise:       1,
					CheckTimeout:    5,
				},
				{
					ID:              "created-0",
					Algorithm:       core.RoundRobinRuleAlgorithm,
					DestinationPort: 199,
					ListenPort:      256,
					Protocol:        core.TCPProtocol,
					CheckEnabled:    true,
					CheckFall:       1,
					CheckInterval:   10,
					CheckProtocol:   core.TCPProtocol,
					CheckRise:       1,
					CheckTimeout:    5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBRController{items: tt.loadBalancerRules}
			lbm := loadBalancerManager{
				loadBalancerRuleController: lbc,
				log:                        logTest.TestLogger{T: t},
			}

			err := lbm.ensureLoadBalancerRules(context.TODO(), tt.service, tt.loadBalancer)
			assert.Equal(t, tt.wantLoadBalancerRules, lbc.items)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_EnsureLoadBalancer(t *testing.T) {
	tests := []struct {
		name string

		loadBalancers []core.LoadBalancer

		service *v1.Service

		wantStatus        *v1.LoadBalancerStatus
		wantLoadBalancers []core.LoadBalancer

		wantErr string
	}{
		{
			name: "uses existing LB",
			loadBalancers: []core.LoadBalancer{
				{
					ID:        "lb_npORVDLVrf7MlghA",
					Name:      "k8s-example-b5216b07-2cb4-4429-8294-23883301a01e",
					IPAddress: &core.IPAddress{Address: "133.7.42.0"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
				Spec:       v1.ServiceSpec{Ports: []v1.ServicePort{}},
			},

			wantStatus: &v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{
				{
					IP: "133.7.42.0",
				},
			}},
			wantLoadBalancers: []core.LoadBalancer{
				{
					ID:        "lb_npORVDLVrf7MlghA",
					Name:      "k8s-example-b5216b07-2cb4-4429-8294-23883301a01e",
					IPAddress: &core.IPAddress{Address: "133.7.42.0"},
				},
			},
		},
		{
			name:          "create lb",
			loadBalancers: []core.LoadBalancer{},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{UID: "b5216b07-2cb4-4429-8294-23883301a01e"},
				Spec:       v1.ServiceSpec{Ports: []v1.ServicePort{}},
			},

			wantStatus: &v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{
				{
					IP: "10.0.0.0",
				},
			}},
			wantLoadBalancers: []core.LoadBalancer{
				{
					ID:           "created-0",
					Name:         "k8s-example-b5216b07-2cb4-4429-8294-23883301a01e",
					IPAddress:    &core.IPAddress{Address: "10.0.0.0"},
					ResourceType: core.TagsResourceType,
					ResourceIDs:  []string{"node-tag-id"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbc := &mockLBController{items: tt.loadBalancers}
			lbrc := &mockLBRController{items: []core.LoadBalancerRule{}}
			lbm := loadBalancerManager{
				config:                     Config{NodeTagID: "node-tag-id"},
				loadBalancerController:     lbc,
				loadBalancerRuleController: lbrc,
				log:                        logTest.TestLogger{T: t},
			}

			status, err := lbm.EnsureLoadBalancer(context.TODO(), "example", tt.service, []*v1.Node{})
			assert.Equal(t, tt.wantLoadBalancers, lbc.items)
			assert.Equal(t, tt.wantStatus, status)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadBalancerManager_UpdateLoadBalancer(t *testing.T) {
	assert.Nil(t,
		(&loadBalancerManager{}).UpdateLoadBalancer(
			context.TODO(),
			"",
			&v1.Service{},
			[]*v1.Node{},
		),
	)
}
