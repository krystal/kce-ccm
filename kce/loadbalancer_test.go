package kce

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

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
