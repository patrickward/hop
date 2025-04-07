# Dispatch Package

The dispatch package provides a simple, type-safe event bus system for Go applications, supporting both synchronous and asynchronous event handling with wildcards and typed payloads. This package is designed for single-binary applications needing simple, type-safe, in-memory event handling. It's ideal for monolithic applications using the Hop framework where events don't need persistence or distributed processing. For distributed systems, message persistence, or advanced features like message routing and transformation, consider using a more comprehensive solution like [Watermill](https://github.com/ThreeDotsLabs/watermill) or a message queue.

## Features

- 🔄 Asynchronous and synchronous event emission
- 🎯 Type-safe payload handling with generics
- 🌟 Wildcard pattern matching for event signatures
- 🛡️ Panic recovery in event handlers
- 📝 Structured logging integration
- 🔍 Context support for cancellation

## Installation

```bash
go get github.com/patrickward/hop/v2/dispatch
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/patrickward/hop/v2/dispatch"
)

func main() {
	// Create a new event dispatcher
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Register an event handler
	dispatcher.On("user.created", func(ctx context.Context, event dispatch.Event) {
		if user, err := dispatch.PayloadAs[User](event); err == nil {
			fmt.Printf("New user created: %s\n", user.Name)
		}
	})

	// Emit an event
	dispatcher.Emit(context.Background(), "user.created", User{
		ID:   "123",
		Name: "John Doe",
	})
}

type User struct {
	ID   string
	Name string
}
```

## Event Signatures

Events use dot-notation signatures (e.g., "user.created", "system.startup"). The signature format is typically:

```
<source>.<event_type>
```

Examples:
- `user.created`
- `order.completed`
- `system.startup`
- `email.sent`

### Wildcard Support

You can use wildcards (*) in event signatures when registering handlers:

```go
// Handle all user events
dispatcher.On("user.*", handler)

// Handle all created events from any source
dispatcher.On("*.created", handler)

// Handle all system events
dispatcher.On("system.*", handler)
```

## Working with Payloads

### Type-Safe Payload Handling

The package provides several helpers for type-safe payload handling:

```go
// Direct type conversion
userEvent, err := dispatch.PayloadAs[User](event)

// Must variant (panics on error)
userEvent := dispatch.MustPayloadAs[User](event)

// Type checking
if dispatch.IsPayloadType[User](event) {
    // Handle user event
}

// Automatic payload handling
dispatcher.On("user.created", dispatch.HandlePayload[User](func(ctx context.Context, user User) {
    fmt.Printf("New user: %s\n", user.Name)
}))
```

### Collection Payloads

Special helpers for common collection types:

```go
// Working with map payloads
config, err := dispatch.PayloadAsMap(event)
regions, err := dispatch.PayloadMapAs[Region](event)

// Working with slice payloads
items, err := dispatch.PayloadAsSlice(event)
users, err := dispatch.PayloadSliceAs[User](event)
```

## Synchronous vs Asynchronous

### Asynchronous Emission (Default)

```go
// Emit events asynchronously (non-blocking)
dispatcher.Emit(ctx, "user.created", user)
```

### Synchronous Emission

```go
// Emit events synchronously (blocks until all handlers complete)
dispatcher.EmitSync(ctx, "user.created", user)
```

## Context Support

All event handlers receive a context.Context, which can be used for cancellation:

```go
dispatcher.On("long.process", func(ctx context.Context, event dispatch.Event) {
    select {
    case <-ctx.Done():
        return // Context cancelled
    case <-time.After(time.Second):
        // Continue processing
    }
})
```

## Error Handling

The dispatcher automatically recovers from panics in event handlers and logs them:

```go
dispatcher.On("risky.operation", func(ctx context.Context, event dispatch.Event) {
    // Even if this panics, other handlers will still run
    panic("something went wrong")
})
```

## Best Practices

1. **Event Naming**: Use consistent naming patterns for events (e.g., `resource.action`)
2. **Type Safety**: Use `PayloadAs` and type-safe handlers where possible
3. **Context Usage**: Pass appropriate contexts for cancellation support
4. **Error Handling**: Always check errors when converting payloads
5. **Documentation**: Document event signatures and their expected payloads

## Thread Safety

The dispatcher is thread-safe and can be safely used from multiple goroutines.

## Performance Considerations

- Async event emission (`Emit`) returns immediately and runs handlers in goroutines
- Sync event emission (`EmitSync`) waits for all handlers to complete
- Wildcard pattern matching adds minimal overhead
- Consider using sync emission for critical path operations where order matters

## Example: Complete System

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/patrickward/hop/v2/dispatch"
)

type OrderCreated struct {
	ID        string
	UserID    string
	Amount    float64
	CreatedAt time.Time
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Register multiple handlers
	dispatcher.On("order.created", dispatch.HandlePayload[OrderCreated](handleNewOrder))
	dispatcher.On("order.*", handleAnyOrderEvent)
	dispatcher.On("*.created", handleAnyCreatedEvent)

	// Emit an event
	order := OrderCreated{
		ID:        "ord_123",
		UserID:    "usr_456",
		Amount:    99.99,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	dispatcher.EmitSync(ctx, "order.created", order)
}

func handleNewOrder(ctx context.Context, order OrderCreated) {
	// Process new order
}

func handleAnyOrderEvent(ctx context.Context, event dispatch.Event) {
	// Handle any order-related event
}

func handleAnyCreatedEvent(ctx context.Context, event dispatch.Event) {
	// Handle any creation event
}
```
