// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.
package main

import (
	"flag"
	"os"

	"github.com/awesomenix/drainsafe/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {

	corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var verbose bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(verbose))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	stopch := ctrl.SetupSignalHandler()

	err = (&controllers.ScheduledEventReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("ScheduledEvent"),
		Recorder: mgr.GetEventRecorderFor("scheduledevent"),
		StopCh:   stopch,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ScheduledEvent")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(stopch); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
