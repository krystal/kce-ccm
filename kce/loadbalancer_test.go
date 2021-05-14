package kce

import (
	"context"
	"fmt"
	"github.com/krystal/go-katapult"
	"github.com/krystal/go-katapult/core"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"testing"
)

type mockLBController struct {
	items []core.LoadBalancer
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
	panic("unimplemented")
	return nil, nil, nil
}

func (lbc *mockLBController) Create(ctx context.Context, org *core.Organization, args *core.LoadBalancerCreateArguments) (*core.LoadBalancer, *katapult.Response, error) {
	panic("unimplemented")
	return nil, nil, nil
}

type mockLBRController struct {
	items []core.LoadBalancerRule
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
	panic("unimplemented")
	return nil, nil, nil
}

func (lbrc *mockLBRController) Update(ctx context.Context, rule *core.LoadBalancerRule, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error) {
	panic("unimplemented")
	return nil, nil, nil
}

func (lbrc *mockLBRController) Create(ctx context.Context, lb *core.LoadBalancer, args core.LoadBalancerRuleArguments) (*core.LoadBalancerRule, *katapult.Response, error) {
	panic("unimplemented")
	return nil, nil, nil
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
			lbm := loadBalancerManager{loadBalancerController: lbc}

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
			lbm := loadBalancerManager{loadBalancerRuleController: lbrc}

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
			lbm := loadBalancerManager{loadBalancerController: lbc}

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
			lbm := loadBalancerManager{loadBalancerController: lbc}

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
			lbm := loadBalancerManager{loadBalancerController: lbc}

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
