/*
Package dispatch provides a lightweight, type-safe event bus implementation for building event-driven
applications in Go. It supports both synchronous and asynchronous event handling, with features
like wildcard pattern matching and typed payload handling through generics.

Basic Usage:

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Register a handler
	dispatcher.On("user.created", func(ctx context.Context, event dispatch.Event) {
	    // Handle the event
	})

	// Emit an event
	dispatcher.Emit(context.Background(), "user.created", userData)

Event Signatures:

Events use dot-notation signatures that typically follow the pattern:

	<source>.<action>

For example:
  - user.created
  - order.completed
  - email.sent

Wildcard pattern matching is supported when registering handlers:
  - "user.*" matches all user events
  - "*.created" matches all creation events
  - "system.*" matches all system events

Type-Safe Payload Handling:

The package provides several helpers for safe payload type conversion:

	// Direct conversion
	user, err := dispatch.PayloadAs[User](event)

	// Type-safe handler
	dispatcher.On("user.created", dispatch.HandlePayload[User](func(ctx context.Context, user User) {
	    // Work with strongly typed user data
	}))

	// Collection helpers
	config, err := dispatch.PayloadAsMap(event)                // For map[string]any
	items, err := dispatch.PayloadAsSlice(event)              // For []any
	regions, err := dispatch.PayloadMapAs[Region](event)      // For map[string]Region
	users, err := dispatch.PayloadSliceAs[User](event)        // For []User

Event Emission:

Events can be emitted either asynchronously (non-blocking) or synchronously (blocking):

	// Async emission (handlers run in goroutines)
	dispatcher.Emit(ctx, "user.created", userData)

	// Sync emission (waits for all handlers to complete)
	dispatcher.EmitSync(ctx, "user.created", userData)

Context Support:

All event handlers receive a context.Context that can be used for cancellation,
timeouts, and passing request-scoped values:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispatcher.EmitSync(ctx, "long.process", data)

Thread Safety:

The event dispatcher is thread-safe and can be safely used from multiple goroutines.
Handler registration and event emission are protected by appropriate synchronization.

When to Use:

This package is designed for single-binary applications needing simple, type-safe, in-memory event handling.
It's ideal for monolithic applications using the Hop framework where events don't need persistence or
distributed processing. For distributed systems, message persistence, or advanced features like message
routing and transformation, consider using a more comprehensive solution like [Watermill](https://github.com/ThreeDotsLabs/watermill)
or a message queue.

SetError Handling:

The event dispatcher automatically recovers from panics in event handlers and logs them
using the provided logger. This ensures that a failing handler won't affect other
handlers or the stability of the event bus.
*/
package dispatch
