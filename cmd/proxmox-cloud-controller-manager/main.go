/*
Copyright 2023 The Kubernetes Authors.

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

// This file should be written by each cloud provider.
// For an minimal working example, please refer to k8s.io/cloud-provider/sample/basic_main.go
// For more details, please refer to k8s.io/kubernetes/cmd/cloud-controller-manager/main.go

// Package main provides the CCM implementation.
package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/sergelogvinov/proxmox-cloud-controller-manager/pkg/proxmox"

	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/names"
	"k8s.io/cloud-provider/options"
	"k8s.io/component-base/cli"
	cliflag "k8s.io/component-base/cli/flag"
	_ "k8s.io/component-base/metrics/prometheus/clientgo" // load all the prometheus client-go plugins
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration
	"k8s.io/klog/v2"
)

func main() {
	ccmOptions, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	fss := cliflag.NamedFlagSets{}
	command := app.NewCloudControllerManagerCommand(ccmOptions, cloudInitializer, app.DefaultInitFuncConstructors, names.CCMControllerAliases(), fss, wait.NeverStop)

	command.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "cloud-provider" {
			if err := flag.Value.Set(proxmox.ProviderName); err != nil {
				klog.Fatalf("unable to set cloud-provider flag value: %s", err)
			}
		}
	})

	code := cli.Run(command)
	os.Exit(code)
}

func cloudInitializer(config *config.CompletedConfig) cloudprovider.Interface {
	cloudConfig := config.ComponentConfig.KubeCloudShared.CloudProvider

	// initialize cloud provider with the cloud provider name and config file provided
	cloud, err := cloudprovider.InitCloudProvider(cloudConfig.Name, cloudConfig.CloudConfigFile)
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
