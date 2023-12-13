/*
Copyright 2023 The KubeArchive Contributors.

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
	goflag "flag"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	flag "github.com/spf13/pflag"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/kubearchive/dynowatch/internal/config"
	"github.com/kubearchive/dynowatch/internal/manager"
	//+kubebuilder:scaffold:imports
)

var (
	scheme    = runtime.NewScheme()
	setupLog  = ctrl.Log.WithName("setup")
	appConfig = config.NewConfig()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme

	appConfig.Init()
}

func main() {

	flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	if err := appConfig.BindPFlag(config.MetricsBindAddressKey, flag.Lookup("metrics-bind-address")); err != nil {
		failNow(err, "Binding flag metrics-bind-address")
	}
	flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	if err := appConfig.BindPFlag(config.HealthzBindAddressKey, flag.Lookup("health-probe-bind-address")); err != nil {
		failNow(err, "Binding flag health-probe-binding-address")
	}
	flag.Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	if err := appConfig.BindPFlag(config.LeaderElectionKey, flag.Lookup("leader-elect")); err != nil {
		failNow(err, "Binding flag leader-elect")
	}
	flag.String("events-source-uri-ref", "locahost",
		"Source of the CloudEvent as a URI reference. The source plus ID of a CloudEvent should be uniquely identifiable.")
	if err := appConfig.BindPFlag(config.CloudEventsSourceURIKey, flag.Lookup("events-source-uri-ref")); err != nil {
		failNow(err, "Binding flag vents-source-uri-ref")
	}
	flag.String("events-target-address", "http://localhost:8082", "The target address to send CloudEvents to.")
	if err := appConfig.BindPFlag(config.CloudEventsTargetAddressKey, flag.Lookup("events-target-address")); err != nil {
		failNow(err, "Binding flag events-target-address")
	}
	opts := &zap.Options{
		Development: true,
	}
	opts.BindFlags(goflag.CommandLine)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if err := appConfig.SafeReadInConfig(); err != nil {
		failNow(err, "Failed to read config")
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(opts)))

	setupLog.Info("Starting dynowatch controller manager")
	setupLog.Info("Configuration data",
		config.MetricsBindAddressKey, appConfig.GetString(config.MetricsBindAddressKey),
		config.HealthzBindAddressKey, appConfig.GetString(config.HealthzBindAddressKey),
		config.LeaderElectionKey, appConfig.GetBool(config.LeaderElectionKey),
		config.CloudEventsSourceURIKey, appConfig.GetString(config.CloudEventsSourceURIKey),
		config.CloudEventsTargetAddressKey, appConfig.GetString(config.CloudEventsTargetAddressKey))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: appConfig.GetString(config.MetricsBindAddressKey)},
		HealthProbeBindAddress: appConfig.GetString(config.HealthzBindAddressKey),
		LeaderElection:         appConfig.GetBool(config.LeaderElectionKey),
		LeaderElectionID:       "fc6a04ff.kubearchive.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		failNow(err, "Unable to start manager")
	}

	protocol, err := cloudevents.NewHTTP()
	if err != nil {
		failNow(err, "Unable to set up cloudevents protocol")
	}
	eventsClient, err := cloudevents.NewClient(protocol, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		failNow(err, "Unable to set up cloudevents client")
	}

	// TODO: Refactor this to its own internal package
	watches, err := appConfig.GetWatches()
	if err != nil {
		failNow(err, "Unable to get watched objects")
	}

	if err := manager.SetupControllers(mgr, eventsClient, watches,
		appConfig.GetString(config.CloudEventsSourceURIKey),
		appConfig.GetString(config.CloudEventsTargetAddressKey)); err != nil {
		failNow(err, "Unable to create controllers")
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		failNow(err, "Unable to set up health check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		failNow(err, "Unable to set up ready check")
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		failNow(err, "Problem running manager")
	}
}

func failNow(err error, msg string) {
	setupLog.Error(err, msg)
	os.Exit(1)
}
