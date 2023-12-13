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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventsclient "github.com/cloudevents/sdk-go/v2/client"
)

// DynamicReconciler reconciles any object with the given GroupVersionKind. When an instance of the
// object is created, updated, or deleted, the reconciler emits a CloudEvent to the configured
// target.
type DynamicReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Name             string
	GroupVersionKind schema.GroupVersionKind
	EventsSource     string
	EventsTarget     string
	EventsClient     cloudeventsclient.Client
}

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update

// Reconcile receives events for the configured object, determines the object's state on the
// cluster, and emits an appropriate CloudEvent. If the event is not delivered, it is retried via
// a requeue.
func (r *DynamicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	eventCtx := cloudevents.ContextWithTarget(ctx, r.EventsTarget)

	obj := r.reconcileTarget()
	event, err := r.newEvent(req.NamespacedName, obj)
	if err != nil {
		log.Error(err, "Failed to create event")
		return ctrl.Result{}, err
	}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		// If not found, return error for requeue
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	result := r.EventsClient.Send(eventCtx, event)
	if cloudevents.IsUndelivered(result) {
		log.Error(result, "Failed to deliver event")
		return ctrl.Result{}, result
	}
	log.Info("Delivered event")

	return ctrl.Result{}, nil
}

// reconcileTarget returns an Unstructured instance of the target object to be reconciled, setting
// the object's GroupVersionKind.
func (r *DynamicReconciler) reconcileTarget() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.GroupVersionKind)
	return obj
}

func (r *DynamicReconciler) newEvent(key types.NamespacedName, obj client.Object) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	event.SetSource(r.EventsSource)
	event.SetType("dynowatch.kubearchive.dev")
	err := event.SetData(cloudevents.ApplicationJSON, map[string]string{
		"kind": obj.GetObjectKind().GroupVersionKind().Kind,
		"apiVersion": fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Group,
			obj.GetObjectKind().GroupVersionKind().Version),
		"namespace": key.Namespace,
		"name":      key.Name,
	})
	return event, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *DynamicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.reconcileTarget()).
		Named(r.Name).
		Complete(r)
}
