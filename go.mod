module github.com/krystal/kce-ccm

go 1.16

require (
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed // indirect
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c // indirect
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cloud-provider v0.21.0
	k8s.io/component-base v0.21.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.1 // indirect
)

// Replace statement fixes issue with older version of etcd and grpc.
// A future version of etcd will fix this. We can remove this statement once
// that has happened.
replace google.golang.org/grpc v1.32.0 => google.golang.org/grpc v1.26.0
