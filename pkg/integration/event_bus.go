// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// EventType represents different types of events in the application
type EventType string

const (
	// EventDataLoaded fired when CSV data is loaded
	EventDataLoaded EventType = "data-loaded"
	// EventPCAStarted fired when PCA analysis starts
	EventPCAStarted EventType = "pca-started"
	// EventPCACompleted fired when PCA analysis completes
	EventPCACompleted EventType = "pca-completed"
	// EventPCAFailed fired when PCA analysis fails
	EventPCAFailed EventType = "pca-failed"
	// EventExportStarted fired when data export starts
	EventExportStarted EventType = "export-started"
	// EventExportCompleted fired when data export completes
	EventExportCompleted EventType = "export-completed"
	// EventAppLaunched fired when companion app is launched
	EventAppLaunched EventType = "app-launched"
	// EventProgressUpdate fired for progress updates
	EventProgressUpdate EventType = "progress-update"
)

// Event represents an application event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// EventHandler is a function that handles events
type EventHandler func(event Event)

// EventBus manages application-wide events
type EventBus struct {
	mu         sync.RWMutex
	handlers   map[EventType][]EventHandler
	history    []Event
	maxHistory int
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers:   make(map[EventType][]EventHandler),
		history:    make([]Event, 0, 100),
		maxHistory: 100,
	}
}

// Subscribe registers a handler for an event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// Return unsubscribe function
	return func() {
		eb.Unsubscribe(eventType, handler)
	}
}

// Unsubscribe removes a handler for an event type
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		// Compare function pointers
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish sends an event to all registered handlers
func (eb *EventBus) Publish(event Event) {
	event.Timestamp = time.Now()

	// Add to history
	eb.mu.Lock()
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.maxHistory {
		eb.history = eb.history[1:]
	}
	handlers := eb.handlers[event.Type]
	eb.mu.Unlock()

	// Call handlers (outside of lock to prevent deadlock)
	for _, handler := range handlers {
		go handler(event)
	}
}

// PublishAsync publishes an event asynchronously with context
func (eb *EventBus) PublishAsync(ctx context.Context, event Event) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			eb.Publish(event)
		}
	}()
}

// GetHistory returns recent events
func (eb *EventBus) GetHistory() []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	history := make([]Event, len(eb.history))
	copy(history, eb.history)
	return history
}

// Clear removes all handlers and history
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers = make(map[EventType][]EventHandler)
	eb.history = make([]Event, 0, eb.maxHistory)
}

// WailsEventAdapter adapts EventBus events to Wails runtime events
type WailsEventAdapter struct {
	eventBus *EventBus
	ctx      context.Context
}

// NewWailsEventAdapter creates a new adapter for Wails events
func NewWailsEventAdapter(ctx context.Context, eventBus *EventBus) *WailsEventAdapter {
	return &WailsEventAdapter{
		eventBus: eventBus,
		ctx:      ctx,
	}
}

// EmitToFrontend converts an Event to JSON and emits it to the frontend
func (adapter *WailsEventAdapter) EmitToFrontend(event Event) error {
	// This would integrate with Wails runtime.EventsEmit
	// For now, we just marshal to JSON as an example
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// In actual implementation, this would call:
	// runtime.EventsEmit(adapter.ctx, string(event.Type), data)
	_ = data // Suppress unused variable warning

	return nil
}

// ProgressTracker tracks progress of long-running operations
type ProgressTracker struct {
	eventBus  *EventBus
	operation string
	total     int
	current   int
	mu        sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(eventBus *EventBus, operation string, total int) *ProgressTracker {
	return &ProgressTracker{
		eventBus:  eventBus,
		operation: operation,
		total:     total,
		current:   0,
	}
}

// Update updates the progress and publishes an event
func (pt *ProgressTracker) Update(current int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.current = current

	percentage := float64(current) / float64(pt.total) * 100

	pt.eventBus.Publish(Event{
		Type: EventProgressUpdate,
		Data: map[string]interface{}{
			"operation":  pt.operation,
			"current":    current,
			"total":      pt.total,
			"percentage": percentage,
		},
	})
}

// Complete marks the operation as complete
func (pt *ProgressTracker) Complete() {
	pt.Update(pt.total)
}
