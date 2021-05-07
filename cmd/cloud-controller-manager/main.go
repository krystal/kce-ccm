/*
Copyright 2021 Krystal Hosting Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// The cloud-controller-manager injects the KCE provider and then hands off
// to the k8s CCM command.
//
// The following is based on guidance based on:
// https://github.com/kubernetes/kubernetes/blob/master/cmd/cloud-controller-manager/main.go
// You may also find the following interesting:
// https://github.com/kubernetes/cloud-provider-gcp/blob/master/cmd/cloud-controller-manager/main.go
// https://github.com/kubernetes/cloud-provider-aws/blob/master/cmd/aws-cloud-controller-manager/main.go

package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/krystal/kce-ccm/kce"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/metrics/prometheus/clientgo" // load all the prometheus client-go plugins
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration
	"k8s.io/klog/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	opts, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	controllerInitializers := app.DefaultInitFuncConstructors
	fss := cliflag.NamedFlagSets{}

	// Create a CCM Command instance
	command := app.NewCloudControllerManagerCommand(opts, cloudInitializer, controllerInitializers, fss, wait.NeverStop)

	// TODO: Switch to utilflag.InitFlags() once k8s switches to Cobra
	// https://github.com/kubernetes/cloud-provider-gcp/issues/215
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func cloudInitializer(config *config.CompletedConfig) cloudprovider.Interface {
	cloudConfig := config.ComponentConfig.KubeCloudShared.CloudProvider

	cloud, err := cloudprovider.InitCloudProvider(kce.ProviderName, cloudConfig.CloudConfigFile)
	if err != nil {
		klog.Fatalf("Cloud provider could not be initialized: %v", err)
	}
	if cloud == nil {
		klog.Fatalf("Cloud provider is nil")
	}

	if !cloud.HasClusterID() {
		if config.ComponentConfig.KubeCloudShared.AllowUntaggedCloud {
			klog.Warning("detected a cluster without a ClusterID.  A ClusterID will be required in the future.  Please tag your cluster to avoid any future issues")
		} else {
			klog.Fatalf("no ClusterID found.  A ClusterID is required for the cloud provider to function properly.  This check can be bypassed by setting the allow-untagged-cloud option")
		}
	}

	return cloud
}
