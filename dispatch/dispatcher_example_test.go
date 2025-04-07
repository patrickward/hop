package dispatch_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/patrickward/hop/v2/dispatch"
)

func ExampleNewEvent() {
	// Create a new event with a payload
	evt := dispatch.NewEvent("user.created", map[string]string{
		"id":    "123",
		"email": "user@example.com",
	})

	fmt.Printf("Signature: %s\n", evt.Signature)
	// Output: Signature: user.created
}

func ExampleDispatcher_basic() {
	// Create a new event dispatcher with a basic logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Register an event handler
	dispatcher.On("user.login", func(ctx context.Context, event dispatch.Event) {
		payload := event.Payload.(map[string]string)
		fmt.Printf("User logged in: %s\n", payload["username"])
	})

	// Emit an event
	dispatcher.Emit(context.Background(), "user.login", map[string]string{
		"username": "alice",
	})

	// Wait for async handler to complete
	time.Sleep(10 * time.Millisecond)
	// Output: User logged in: alice
}

func ExampleDispatcher_wildcards() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	var mu sync.Mutex
	var results []string

	// Register handlers with wildcards
	dispatcher.On("user.*", func(ctx context.Context, event dispatch.Event) {
		mu.Lock()
		results = append(results, fmt.Sprintf("User event: %s", event.Signature))
		mu.Unlock()
	})

	dispatcher.On("*.created", func(ctx context.Context, event dispatch.Event) {
		mu.Lock()
		results = append(results, fmt.Sprintf("Created event: %s", event.Signature))
		mu.Unlock()
	})

	// Emit events synchronously
	dispatcher.EmitSync(context.Background(), "user.created", nil)
	dispatcher.EmitSync(context.Background(), "user.deleted", nil)
	dispatcher.EmitSync(context.Background(), "post.created", nil)

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

func ExampleDispatcher_syncEmit() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Register event handlers
	dispatcher.On("task.process", func(ctx context.Context, event dispatch.Event) {
		fmt.Println("Processing task...")
		time.Sleep(10 * time.Millisecond)
		fmt.Println("Task completed")
	})

	// EmitSync will wait for all handlers to complete
	fmt.Println("Starting task")
	dispatcher.EmitSync(context.Background(), "task.process", nil)
	fmt.Println("All processing complete")

	// Output:
	// Starting task
	// Processing task...
	// Task completed
	// All processing complete
}

func ExampleDispatcher_contextCancellation() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	// Create a context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Register a handler that respects context cancellation
	dispatcher.On("long.task", func(ctx context.Context, event dispatch.Event) {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled")
			return
		case <-time.After(100 * time.Millisecond):
			fmt.Println("Task completed")
		}
	})

	// Emit event with cancellable context
	dispatcher.EmitSync(ctx, "long.task", nil)
	// Output: Task cancelled
}

func ExampleDispatcher_multipleHandlers() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatch.NewDispatcher(logger)

	var mu sync.Mutex
	var results []string

	// Register multiple handlers for the same event
	dispatcher.On("notification.sent", func(ctx context.Context, event dispatch.Event) {
		mu.Lock()
		results = append(results, "Logging notification")
		mu.Unlock()
	})

	dispatcher.On("notification.sent", func(ctx context.Context, event dispatch.Event) {
		mu.Lock()
		results = append(results, "Sending analytics")
		mu.Unlock()
	})

	dispatcher.On("notification.sent", func(ctx context.Context, event dispatch.Event) {
		mu.Lock()
		results = append(results, "Updating cache")
		mu.Unlock()
	})

	// Emit event synchronously - all handlers will be called
	dispatcher.EmitSync(context.Background(), "notification.sent", nil)

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
