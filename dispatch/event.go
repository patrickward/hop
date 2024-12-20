package dispatch

import (
	"context"
	"time"
)

// Event represents a system event with a simplified structure
type Event struct {
	ID        string    `json:"id"`
	Signature string    `json:"signature"` // e.g. "hop.system.start"
	Payload   any       `json:"payload,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Handler processes an event
type Handler func(ctx context.Context, event Event)

// NewEvent creates an event with the given signature and optional payload
func NewEvent(signature string, payload any) Event {
	return Event{
		ID:        generateEventID(),
		Signature: signature,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}
