package dispatch_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/patrickward/hop/dispatch"
)

func ExamplePayloadAs() {
	// Example event with a structured payload
	type UserCreated struct {
		ID   string
		Name string
	}

	evt := dispatch.NewEvent("user.created", UserCreated{
		ID:   "123",
		Name: "John Doe",
	})

	// Safe conversion with error handling
	user, err := dispatch.PayloadAs[UserCreated](evt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User created: %s\n", user.Name)
	// Output: User created: John Doe
}

func ExampleHandlePayload() {
	type UserCreated struct {
		ID   string
		Name string
	}

	logger := newTestLogger(os.Stderr)
	// Create a test event bus
	bus := dispatch.NewDispatcher(logger) // You'd normally pass a logger here

	// Register handler with automatic payload conversion
	bus.On("user.created", dispatch.HandlePayload[UserCreated](func(ctx context.Context, user UserCreated) {
		fmt.Printf("Processing user: %s\n", user.Name)
	}))

	// Emit an event
	ctx := context.Background()
	bus.EmitSync(ctx, "user.created", UserCreated{
		ID:   "123",
		Name: "John Doe",
	})
	// Output: Processing user: John Doe
}

func ExamplePayloadAsMap() {
	// Create an event with a map payload
	evt := dispatch.NewEvent("config.updated", map[string]any{
		"database": "postgres",
		"port":     5432,
	})

	// Use the convenience function for map payloads
	config, err := dispatch.PayloadAsMap(evt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if dbName, ok := config["database"].(string); ok {
		fmt.Printf("Database: %s\n", dbName)
	}
	// Output: Database: postgres
}

func ExamplePayloadAsSlice() {
	// Create an event with a slice payload
	evt := dispatch.NewEvent("users.updated", []any{
		"john",
		"jane",
	})

	// Use the convenience function for slice payloads
	users, err := dispatch.PayloadAsSlice(evt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Users: %v\n", users)

	// Output: Users: [john jane]
}

func ExamplePayloadMapAs() {
	// Define a type we want to convert to
	type Region struct {
		Code string
		Name string
	}

	var mu sync.Mutex
	var results []string

	// Create an event with a map payload
	evt := dispatch.NewEvent("regions.updated", map[string]any{
		"us-east": Region{Code: "USE", Name: "US East"},
		"us-west": Region{Code: "USW", Name: "US West"},
	})

	// Convert the payload to a map of Regions
	regions, err := dispatch.PayloadMapAs[Region](evt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Process the typed regions
	for key, region := range regions {
		mu.Lock()
		results = append(results, fmt.Sprintf("Region %s: %s (%s)", key, region.Name, region.Code))
		mu.Unlock()
	}

	// Sort and print results for consistent output
	sort.Strings(results)
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// Region us-east: US East (USE)
	// Region us-west: US West (USW)
}

func ExamplePayloadSliceAs() {
	// Define a type we want to convert to
	type User struct {
		ID   string
		Name string
	}

	// Create an event with a slice payload
	evt := dispatch.NewEvent("users.imported", []any{
		User{ID: "1", Name: "Alice"},
		User{ID: "2", Name: "Bob"},
	})

	// Convert the payload to a slice of Users
	users, err := dispatch.PayloadSliceAs[User](evt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Process the typed users
	for _, user := range users {
		fmt.Printf("User: %s (ID: %s)\n", user.Name, user.ID)
	}
	// Output:
	// User: Alice (ID: 1)
	// User: Bob (ID: 2)
}
