package hop

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var eventID atomic.Uint64

// Event represents a system event with a simplified structure
type Event struct {
	ID        string    `json:"id"`
	Signature string    `json:"signature"` // e.g. "hop.system.start"
	Payload   any       `json:"payload,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EventHandler processes an event
type EventHandler func(ctx context.Context, event Event)

// EventBus manages event publishing and subscription
type EventBus struct {
	handlers map[string][]EventHandler // key is the event signature
	logger   *slog.Logger
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus(logger *slog.Logger) *EventBus {
	if logger == nil {
		panic("logger is required for event bus")
	}

	return &EventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

// NewEvent creates an event with the given signature and optional payload
func NewEvent(signature string, payload any) Event {
	return Event{
		ID:        generateEventID(),
		Signature: signature,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// On registers a handler for an event signature
// Supports wildcards: "hop.*" or "*.system.start"
func (eb *EventBus) On(signature string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.handlers[signature] == nil {
		eb.handlers[signature] = []EventHandler{}
	}
	eb.handlers[signature] = append(eb.handlers[signature], handler)

	source, eventType := parseSignature(signature)
	eb.logger.Debug("event handler registered",
		slog.String("signature", signature),
		slog.String("source", source),
		slog.String("type", eventType))
}

// Emit sends an event to all registered handlers asynchronously
func (eb *EventBus) Emit(ctx context.Context, signature string, payload any) {
	event := NewEvent(signature, payload)
	eb.mu.RLock()
	var matchingHandlers []EventHandler
	for pattern, handlers := range eb.handlers {
		if matchSignature(pattern, event.Signature) {
			matchingHandlers = append(matchingHandlers, handlers...)
		}
	}
	eb.mu.RUnlock()

	source, eventType := parseSignature(event.Signature)
	eb.logger.Debug("emitting event",
		slog.String("signature", event.Signature),
		slog.String("source", source),
		slog.String("type", eventType))

	if len(matchingHandlers) == 0 {
		eb.logger.Debug("no handlers for event",
			slog.String("signature", event.Signature))
		return
	}

	for _, handler := range matchingHandlers {
		h := handler // Capture handler for goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					eb.logger.Error("panic in event handler",
						slog.Any("panic", r),
						slog.String("signature", event.Signature))
				}
			}()

			h(ctx, event)
		}()
	}
}

// EmitSync sends an event and waits for all handlers to complete
func (eb *EventBus) EmitSync(ctx context.Context, signature string, payload any) {
	event := NewEvent(signature, payload)
	eb.mu.RLock()
	var matchingHandlers []EventHandler
	for pattern, handlers := range eb.handlers {
		if matchSignature(pattern, event.Signature) {
			matchingHandlers = append(matchingHandlers, handlers...)
		}
	}
	eb.mu.RUnlock()

	if len(matchingHandlers) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(matchingHandlers))

	for _, handler := range matchingHandlers {
		h := handler
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					eb.logger.Error("panic in event handler",
						slog.Any("panic", r),
						slog.String("signature", event.Signature))
				}
			}()

			h(ctx, event)
		}()
	}

	wg.Wait()
}

// parseSignature splits a signature into source and event type
func parseSignature(signature string) (source, eventType string) {
	parts := strings.SplitN(signature, ".", 2)
	if len(parts) < 2 {
		return "unknown", signature
	}
	return parts[0], parts[1]
}

// matchSignature checks if a pattern matches a signature
func matchSignature(pattern, signature string) bool {
	if pattern == "*" {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	signatureParts := strings.Split(signature, ".")

	if len(patternParts) != len(signatureParts) {
		return false
	}

	for i, part := range patternParts {
		if part != "*" && part != signatureParts[i] {
			return false
		}
	}

	return true
}

// generateEventID creates a unique event ID
func generateEventID() string {
	id := eventID.Add(1)
	return fmt.Sprintf("evt_%d", id)
}
