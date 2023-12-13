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

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventsclient "github.com/cloudevents/sdk-go/v2/client"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	EventsSource string
	EventsTarget string
	EventsClient cloudeventsclient.Client
}

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Job object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	eventCtx := cloudevents.ContextWithTarget(ctx, r.EventsTarget)

	job := &batchv1.Job{}
	event := r.newEvent(req.NamespacedName)
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		// If not found, return error for requeue
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	result := r.EventsClient.Send(eventCtx, event)
	if cloudevents.IsUndelivered(result) {
		log.Error(result, "failed to deliver event")
		return ctrl.Result{}, result
	}
	log.Info("delivered event")

	return ctrl.Result{}, nil
}

func (r *JobReconciler) newEvent(key types.NamespacedName) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource(r.EventsSource)
	event.SetType("dynowatch.kubearchive.dev")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{
		"kind":       "Job",
		"apiVersion": "batch/v1",
		"namespace":  key.Namespace,
		"name":       key.Name,
	})
	return event
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).
		Complete(r)
}
