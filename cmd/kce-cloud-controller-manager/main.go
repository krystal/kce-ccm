/*
Copyright 2016 The Kubernetes Authors.
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

// The external controller manager is responsible for running controller loops that
// are cloud provider dependent. It uses the API to listen to new events on resources.

// This file should be written by each cloud provider.
// For an minimal working example, please refer to k8s.io/cloud-provider/sample/basic_main.go
// For an advanced example, please refer to k8s.io/cloud-provider/sample/advanced_main.go
// For more details, please refer to k8s.io/kubernetes/cmd/cloud-controller-manager/main.go
// The current file demonstrate how other cloud provider should leverage CCM and it uses fake parameters. Please modify for your own use.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/krystal/kceccm/kce"
	"github.com/spf13/pflag"

	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/options"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/metrics/prometheus/clientgo" // load all the prometheus client-go plugins
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration
	"k8s.io/klog/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	s, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	fs := pflag.NewFlagSet("kce-cloud-controller-manager", pflag.ContinueOnError)
	fs.SetNormalizeFunc(flag.WordSepNormalizeFunc)
	for _, f := range s.Flags([]string{"cloud-node", "cloud-node-lifecycle", "service", "route"}, app.ControllersDisabledByDefault.List()).FlagSets {
		fs.AddFlagSet(f)
	}

	err = fs.Parse(os.Args[1:])
	if err != nil {
		if err != pflag.ErrHelp {
			klog.Errorf("could not parse arguments: %v", err)
		}
		os.Exit(1)
	}

	c, err := s.Config([]string{}, app.ControllersDisabledByDefault.List())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	cloud, err := cloudprovider.InitCloudProvider(kce.ProviderName, "")
	if err != nil {
		klog.Fatalf("Cloud provider could not be initialized: %v", err)
	}
	if cloud == nil {
		klog.Fatalf("cloud provider is nil")
	}

	if !cloud.HasClusterID() {
		if c.ComponentConfig.KubeCloudShared.AllowUntaggedCloud {
			klog.Warning("detected a cluster without a ClusterID.  A ClusterID will be required in the future.  Please tag your cluster to avoid any future issues")
		} else {
			klog.Fatalf("no ClusterID found.  A ClusterID is required for the cloud provider to function properly.  This check can be bypassed by setting the allow-untagged-cloud option")
		}
	}

	// Initialize the cloud provider with a reference to the clientBuilder
	cloud.Initialize(c.ClientBuilder, make(chan struct{}))
	// Set the informer on the user cloud object
	if informerUserCloud, ok := cloud.(cloudprovider.InformerUser); ok {
		informerUserCloud.SetInformers(c.SharedInformers)
	}

	controllerInitializers := app.DefaultControllerInitializers(c.Complete(), cloud)
	command := app.NewCloudControllerManagerCommand(s, c, controllerInitializers)

	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.
	// Here is an sample
	pflag.CommandLine.SetNormalizeFunc(flag.WordSepNormalizeFunc)
	// utilflag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	// the flags could be set before execute
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "cloud-provider" {
			flag.Value.Set(kce.ProviderName)
			return
		}
	})

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
