// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package integration

import (
	"context"
	"testing"
	"time"
)

func TestEventBus_NewEventBus(t *testing.T) {
	bus := NewEventBus()

	if bus == nil {
		t.Fatal("NewEventBus() returned nil")
	}
}

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()
	received := make(chan bool, 1)

	bus.Subscribe(EventDataLoaded, func(event Event) {
		if event.Type != EventDataLoaded {
			t.Errorf("expected type %v, got %v", EventDataLoaded, event.Type)
		}
		received <- true
	})

	bus.Publish(Event{Type: EventDataLoaded, Data: map[string]interface{}{"test": "data"}})

	select {
	case <-received:
		// Success - event was received
	case <-time.After(time.Second):
		t.Error("timeout: event was not received by subscriber")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	received := make(chan struct{}, 2)

	bus.Subscribe(EventDataLoaded, func(event Event) {
		received <- struct{}{}
	})

	bus.Subscribe(EventDataLoaded, func(event Event) {
		received <- struct{}{}
	})

	bus.Publish(Event{Type: EventDataLoaded})

	// Wait for both subscribers to receive the event
	timeout := time.After(time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-received:
			// Success - subscriber received event
		case <-timeout:
			t.Errorf("timeout: only %d of 2 subscribers received event", i)
			return
		}
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	received := make(chan bool, 1)

	handler := func(event Event) {
		received <- true
	}

	unsubscribe := bus.Subscribe(EventDataLoaded, handler)
	unsubscribe()

	bus.Publish(Event{Type: EventDataLoaded})

	// Should timeout since handler was unsubscribed
	select {
	case <-received:
		t.Error("unsubscribed handler should not receive events")
	case <-time.After(100 * time.Millisecond):
		// Success - handler did not receive event (timeout expected)
	}
}

func TestEventBus_PublishAsync(t *testing.T) {
	bus := NewEventBus()
	received := make(chan bool, 1)

	bus.Subscribe(EventPCAStarted, func(event Event) {
		time.Sleep(10 * time.Millisecond)
		received <- true
	})

	ctx := context.Background()
	bus.PublishAsync(ctx, Event{Type: EventPCAStarted})

	select {
	case <-received:
		// Success - async event was received
	case <-time.After(time.Second):
		t.Error("timeout: async event was not received")
	}
}

func TestEventBus_GetHistory(t *testing.T) {
	bus := NewEventBus()

	bus.Publish(Event{Type: EventDataLoaded})
	bus.Publish(Event{Type: EventPCACompleted})

	time.Sleep(50 * time.Millisecond)

	history := bus.GetHistory()

	if len(history) < 2 {
		t.Errorf("expected at least 2 events in history, got %d", len(history))
	}
}

func TestEventBus_Clear(t *testing.T) {
	bus := NewEventBus()

	bus.Subscribe(EventDataLoaded, func(event Event) {})
	bus.Publish(Event{Type: EventDataLoaded})

	time.Sleep(50 * time.Millisecond)

	bus.Clear()

	history := bus.GetHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history after clear, got %d events", len(history))
	}
}

func TestProgressTracker_NewProgressTracker(t *testing.T) {
	bus := NewEventBus()
	tracker := NewProgressTracker(bus, "test-task", 10)

	if tracker == nil {
		t.Fatal("NewProgressTracker() returned nil")
	}
}

func TestProgressTracker_Update(t *testing.T) {
	bus := NewEventBus()
	tracker := NewProgressTracker(bus, "test-task", 10)

	received := make(chan bool, 1)
	bus.Subscribe(EventProgressUpdate, func(event Event) {
		received <- true
	})

	tracker.Update(5)

	select {
	case <-received:
		// Success - progress update event received
	case <-time.After(time.Second):
		t.Error("timeout: progress update event not received")
	}
}

func TestProgressTracker_Complete(t *testing.T) {
	bus := NewEventBus()
	tracker := NewProgressTracker(bus, "test-task", 10)

	tracker.Complete()

	time.Sleep(50 * time.Millisecond)

	history := bus.GetHistory()
	if len(history) == 0 {
		t.Error("no events in history after complete")
	}
}

// ─── WailsEventAdapter ────────────────────────────────────────────────────────

func TestWailsEventAdapter_New(t *testing.T) {
	bus := NewEventBus()
	ctx := context.Background()
	adapter := NewWailsEventAdapter(ctx, bus)
	if adapter == nil {
		t.Fatal("NewWailsEventAdapter returned nil")
	}
}

func TestWailsEventAdapter_EmitToFrontend(t *testing.T) {
	bus := NewEventBus()
	ctx := context.Background()
	adapter := NewWailsEventAdapter(ctx, bus)

	event := Event{
		Type: EventDataLoaded,
		Data: map[string]interface{}{"rows": 100},
	}
	if err := adapter.EmitToFrontend(event); err != nil {
		t.Errorf("EmitToFrontend returned unexpected error: %v", err)
	}
}
