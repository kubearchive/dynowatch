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

package manager

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubearchive/dynowatch/internal/config"
	"github.com/kubearchive/dynowatch/internal/controller"
)

var log = ctrl.Log.WithName("manager")

func SetupControllers(mgr manager.Manager, client cloudevents.Client, watches []config.Watch, eventsSource string, eventsTarget string) error {
	for _, watchObj := range watches {
		gvk := schema.GroupVersionKind{
			Group:   watchObj.Group,
			Version: watchObj.Version,
			Kind:    watchObj.Kind,
		}
		reconciler := &controller.DynamicReconciler{
			Client:           mgr.GetClient(),
			Scheme:           mgr.GetScheme(),
			Name:             watchObj.Name,
			GroupVersionKind: gvk,
			EventsSource:     eventsSource,
			EventsTarget:     eventsTarget,
			EventsClient:     client,
		}
		if err := reconciler.SetupWithManager(mgr); err != nil {
			return err
		}
		log.Info("Setup controller", "controller", watchObj.Name, "controllerGroup", watchObj.Group, "controllerKind", watchObj.Kind)
	}
	return nil
}
