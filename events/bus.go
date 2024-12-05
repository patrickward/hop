package events

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
)

var eventID atomic.Uint64

// Bus manages event publishing and subscription
type Bus struct {
	handlers map[string][]Handler // key is the event signature
	logger   *slog.Logger
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus(logger *slog.Logger) *Bus {
	if logger == nil {
		panic("logger is required for event bus")
	}

	return &Bus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

// On registers a handler for an event signature
// Supports wildcards: "hop.*" or "*.system.start"
func (b *Bus) On(signature string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.handlers[signature] == nil {
		b.handlers[signature] = []Handler{}
	}
	b.handlers[signature] = append(b.handlers[signature], handler)

	source, eventType := parseSignature(signature)
	b.logger.Debug("event handler registered",
		slog.String("signature", signature),
		slog.String("source", source),
		slog.String("type", eventType))
}

// Emit sends an event to all registered handlers asynchronously
func (b *Bus) Emit(ctx context.Context, signature string, payload any) {
	event := NewEvent(signature, payload)
	b.mu.RLock()
	var matchingHandlers []Handler
	for pattern, handlers := range b.handlers {
		if matchSignature(pattern, event.Signature) {
			matchingHandlers = append(matchingHandlers, handlers...)
		}
	}
	b.mu.RUnlock()

	source, eventType := parseSignature(event.Signature)
	b.logger.Debug("emitting event",
		slog.String("signature", event.Signature),
		slog.String("source", source),
		slog.String("type", eventType))

	if len(matchingHandlers) == 0 {
		b.logger.Debug("no handlers for event",
			slog.String("signature", event.Signature))
		return
	}

	for _, handler := range matchingHandlers {
		h := handler // Capture handler for goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Error("panic in event handler",
						slog.Any("panic", r),
						slog.String("signature", event.Signature))
				}
			}()

			h(ctx, event)
		}()
	}
}

// EmitSync sends an event and waits for all handlers to complete
func (b *Bus) EmitSync(ctx context.Context, signature string, payload any) {
	event := NewEvent(signature, payload)
	b.mu.RLock()
	var matchingHandlers []Handler
	for pattern, handlers := range b.handlers {
		if matchSignature(pattern, event.Signature) {
			matchingHandlers = append(matchingHandlers, handlers...)
		}
	}
	b.mu.RUnlock()

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
					b.logger.Error("panic in event handler",
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
