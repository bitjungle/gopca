// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
	received := false

	bus.Subscribe(EventDataLoaded, func(event Event) {
		received = true
		if event.Type != EventDataLoaded {
			t.Errorf("expected type %v, got %v", EventDataLoaded, event.Type)
		}
	})

	bus.Publish(Event{Type: EventDataLoaded, Data: map[string]interface{}{"test": "data"}})

	time.Sleep(50 * time.Millisecond)

	if !received {
		t.Error("event was not received by subscriber")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	count := 0

	bus.Subscribe(EventDataLoaded, func(event Event) {
		count++
	})

	bus.Subscribe(EventDataLoaded, func(event Event) {
		count++
	})

	bus.Publish(Event{Type: EventDataLoaded})

	time.Sleep(50 * time.Millisecond)

	if count != 2 {
		t.Errorf("expected 2 callbacks, got %d", count)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	received := false

	handler := func(event Event) {
		received = true
	}

	unsubscribe := bus.Subscribe(EventDataLoaded, handler)
	unsubscribe()

	bus.Publish(Event{Type: EventDataLoaded})

	time.Sleep(50 * time.Millisecond)

	if received {
		t.Error("unsubscribed handler should not receive events")
	}
}

func TestEventBus_PublishAsync(t *testing.T) {
	bus := NewEventBus()
	received := false

	bus.Subscribe(EventPCAStarted, func(event Event) {
		time.Sleep(10 * time.Millisecond)
		received = true
	})

	ctx := context.Background()
	bus.PublishAsync(ctx, Event{Type: EventPCAStarted})

	time.Sleep(100 * time.Millisecond)

	if !received {
		t.Error("async event was not received")
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

	received := false
	bus.Subscribe(EventProgressUpdate, func(event Event) {
		received = true
	})

	tracker.Update(5)

	time.Sleep(50 * time.Millisecond)

	if !received {
		t.Error("progress update event not received")
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
