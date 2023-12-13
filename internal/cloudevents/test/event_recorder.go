package test

import (
	"sync"
	"sync/atomic"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type EventRecorder struct {
	recordLock sync.Mutex
	recording  *atomic.Bool
	events     []cloudevents.Event
}

func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		events:    []cloudevents.Event{},
		recording: &atomic.Bool{},
	}
}

func (e *EventRecorder) Start() {
	e.recording.Store(true)
}

func (e *EventRecorder) Stop() {
	e.recording.Store(false)
}

func (e *EventRecorder) Record(event cloudevents.Event) {
	if !e.recording.Load() {
		return
	}
	e.recordLock.Lock()
	defer e.recordLock.Unlock()
	e.events = append(e.events, event)

}

func (e *EventRecorder) Events() []cloudevents.Event {
	e.recordLock.Lock()
	defer e.recordLock.Unlock()
	events := append([]cloudevents.Event{}, e.events...)
	return events
}

func (e *EventRecorder) Clear() {
	e.recordLock.Lock()
	defer e.recordLock.Unlock()
	e.events = []cloudevents.Event{}
}
