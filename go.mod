module github.com/krystal/kce-ccm

go 1.16

require (
	github.com/krystal/go-katapult v0.0.0-20210506080300-04e5ede56f9d
	github.com/sethvargo/go-envconfig v0.3.5
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cloud-provider v0.21.0
	k8s.io/component-base v0.21.0
	k8s.io/klog/v2 v2.8.0
)

// Replace statement fixes issue with older version of etcd and grpc.
// A future version of etcd will fix this. We can remove this statement once
// that has happened.
replace google.golang.org/grpc => google.golang.org/grpc v1.27.0
