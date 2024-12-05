package events_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/patrickward/hop/events"
)

func ExampleNewEvent() {
	// Create a new event with a payload
	evt := events.NewEvent("user.created", map[string]string{
		"id":    "123",
		"email": "user@example.com",
	})

	fmt.Printf("Signature: %s\n", evt.Signature)
	// Output: Signature: user.created
}

func ExampleBus_basic() {
	// Create a new event bus with a basic logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bus := events.NewEventBus(logger)

	// Register an event handler
	bus.On("user.login", func(ctx context.Context, event events.Event) {
		payload := event.Payload.(map[string]string)
		fmt.Printf("User logged in: %s\n", payload["username"])
	})

	// Emit an event
	bus.Emit(context.Background(), "user.login", map[string]string{
		"username": "alice",
	})

	// Wait for async handler to complete
	time.Sleep(10 * time.Millisecond)
	// Output: User logged in: alice
}

func ExampleBus_wildcards() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bus := events.NewEventBus(logger)

	var mu sync.Mutex
	var results []string

	// Register handlers with wildcards
	bus.On("user.*", func(ctx context.Context, event events.Event) {
		mu.Lock()
		results = append(results, fmt.Sprintf("User event: %s", event.Signature))
		mu.Unlock()
	})

	bus.On("*.created", func(ctx context.Context, event events.Event) {
		mu.Lock()
		results = append(results, fmt.Sprintf("Created event: %s", event.Signature))
		mu.Unlock()
	})

	// Emit events synchronously
	bus.EmitSync(context.Background(), "user.created", nil)
	bus.EmitSync(context.Background(), "user.deleted", nil)
	bus.EmitSync(context.Background(), "post.created", nil)

	// Sort and print results
	sort.Strings(results)
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// Created event: post.created
	// Created event: user.created
	// User event: user.created
	// User event: user.deleted
}

func ExampleBus_syncEmit() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bus := events.NewEventBus(logger)

	// Register event handlers
	bus.On("task.process", func(ctx context.Context, event events.Event) {
		fmt.Println("Processing task...")
		time.Sleep(10 * time.Millisecond)
		fmt.Println("Task completed")
	})

	// EmitSync will wait for all handlers to complete
	fmt.Println("Starting task")
	bus.EmitSync(context.Background(), "task.process", nil)
	fmt.Println("All processing complete")

	// Output:
	// Starting task
	// Processing task...
	// Task completed
	// All processing complete
}

func ExampleBus_contextCancellation() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bus := events.NewEventBus(logger)

	// Create a context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Register a handler that respects context cancellation
	bus.On("long.task", func(ctx context.Context, event events.Event) {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled")
			return
		case <-time.After(100 * time.Millisecond):
			fmt.Println("Task completed")
		}
	})

	// Emit event with cancellable context
	bus.EmitSync(ctx, "long.task", nil)
	// Output: Task cancelled
}

func ExampleBus_multipleHandlers() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bus := events.NewEventBus(logger)

	var mu sync.Mutex
	var results []string

	// Register multiple handlers for the same event
	bus.On("notification.sent", func(ctx context.Context, event events.Event) {
		mu.Lock()
		results = append(results, "Logging notification")
		mu.Unlock()
	})

	bus.On("notification.sent", func(ctx context.Context, event events.Event) {
		mu.Lock()
		results = append(results, "Sending analytics")
		mu.Unlock()
	})

	bus.On("notification.sent", func(ctx context.Context, event events.Event) {
		mu.Lock()
		results = append(results, "Updating cache")
		mu.Unlock()
	})

	// Emit event synchronously - all handlers will be called
	bus.EmitSync(context.Background(), "notification.sent", nil)

	// Sort and print results
	sort.Strings(results)
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// Logging notification
	// Sending analytics
	// Updating cache
}
