/*
Copyright 2021.

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

package main

import (
	"flag"
	"os"

	//+kubebuilder:scaffold:imports

	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
	"github.com/thoth-station/meteor-operator/controllers/cnbi"
	common "github.com/thoth-station/meteor-operator/controllers/common"
	meteor "github.com/thoth-station/meteor-operator/controllers/meteor"
	shower "github.com/thoth-station/meteor-operator/controllers/shower"
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	ctrlConfig = meteorv1alpha1.MeteorConfig{}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(meteorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(pipelinev1beta1.AddToScheme(scheme))
	if ctrlConfig.Spec.EnableShower {
		utilruntime.Must(routev1.AddToScheme(scheme))
	}
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	common.InitMetrics()
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	var err error

	options := ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      "127.0.0.1:8080",
		Port:                    9443,
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          true,
		LeaderElectionID:        "05b1bff9.meteor.zone",
		LeaderElectionNamespace: "aicoe-meteor",
	}

	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))

		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&meteor.MeteorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Meteor")
		os.Exit(1)
	}

	if ctrlConfig.Spec.EnableShower {
		if err = (&shower.ShowerReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Shower")
			os.Exit(1)
		}
	}

	if err = (&cnbi.CustomRuntimeEnvironmentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomRuntimeEnvironment")
		os.Exit(1)
	}

	/* Since we might want to run
	   the webhooks separately, or not run them when testing our controller
	   locally, we'll put them behind an environment variable.
	   We'll just make sure to set `ENABLE_WEBHOOKS=false` when we run locally.
	*/
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&meteorv1alpha1.CustomRuntimeEnvironment{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CustomRuntimeEnvironment")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
