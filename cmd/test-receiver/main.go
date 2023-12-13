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
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/kubearchive/dynowatch/internal/cloudevents/test"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var mainLog = ctrl.Log.WithName("main")

func main() {
	var addr string

	flag.StringVar(&addr, "address", ":8080", "The address to listen to.")
	opts := &zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(opts)))

	ctx := ctrl.SetupSignalHandler()
	handler, err := test.NewReceiverHTTPHandler(ctx, nil)
	if err != nil {
		mainLog.Error(err, "failed to create CloudEventHandler")
		os.Exit(1)
	}
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	mainLog = mainLog.WithValues("address", addr)
	mainLog.Info("running cloudEvent receiver")
	go func() {
		if err := server.ListenAndServe(); !errors.Is(http.ErrServerClosed, err) {
			mainLog.Error(err, "error running HTTP receiver")
			os.Exit(1)
		}
	}()
	mainLog.Info("cloudEvent receiver started")
	<-ctx.Done()
	mainLog.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		mainLog.Error(err, "failed to shut down receiver")
	}
	mainLog.Info("shutdown complete")
}
