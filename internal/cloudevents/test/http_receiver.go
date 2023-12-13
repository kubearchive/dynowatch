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

package test

import (
	"context"
	"net/http/httptest"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/go-logr/logr"
	ctrlLog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log logr.Logger

type TestReceiver struct {
	*httptest.Server
	eventRecorder *EventRecorder
}

func NewTestReceiver(ctx context.Context) (*TestReceiver, error) {
	recorder := NewEventRecorder()
	handler, err := NewReceiverHTTPHandler(ctx, recorder)
	if err != nil {
		return nil, err
	}
	server := httptest.NewUnstartedServer(handler)
	return &TestReceiver{
		Server:        server,
		eventRecorder: recorder,
	}, nil
}

func NewReceiverHTTPHandler(ctx context.Context, recorder *EventRecorder) (*client.EventReceiver, error) {
	protocol, err := cloudevents.NewHTTP()
	if err != nil {
		return nil, err
	}
	log = ctrlLog.Log.WithName("handler")
	if recorder == nil {
		recorder = NewEventRecorder()
	}
	eventHander := newEventHandler(recorder)
	handler, err := cloudevents.NewHTTPReceiveHandler(ctx, protocol, eventHander.handleEvent)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func (r *TestReceiver) StartRecorder() {
	r.eventRecorder.Start()
}

func (r *TestReceiver) StopRecorder() {
	r.eventRecorder.Stop()
}

func (r *TestReceiver) ClearEvents() {
	r.eventRecorder.Clear()
}

func (r *TestReceiver) GetEvents() []cloudevents.Event {
	return r.eventRecorder.Events()
}

type eventHandler struct {
	recorder *EventRecorder
}

func newEventHandler(recorder *EventRecorder) *eventHandler {
	return &eventHandler{
		recorder: recorder,
	}
}

func (e *eventHandler) handleEvent(ctx context.Context, event cloudevents.Event) {
	log.Info("received event", "event", event)
	e.recorder.Record(event)
}
