package hop_test

import (
	"context"
	"fmt"
	"os"

	"github.com/patrickward/hop"
)

func ExamplePayloadAs() {
	// Example event with a structured payload
	type UserCreated struct {
		ID   string
		Name string
	}

	event := hop.NewEvent("user.created", UserCreated{
		ID:   "123",
		Name: "John Doe",
	})

	// Safe conversion with error handling
	user, err := hop.PayloadAs[UserCreated](event)
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
	bus := hop.NewEventBus(logger) // You'd normally pass a logger here

	// Register handler with automatic payload conversion
	bus.On("user.created", hop.HandlePayload[UserCreated](func(ctx context.Context, user UserCreated) {
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
	event := hop.NewEvent("config.updated", map[string]any{
		"database": "postgres",
		"port":     5432,
	})

	// Use the convenience function for map payloads
	config, err := hop.PayloadAsMap(event)
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
	event := hop.NewEvent("users.updated", []any{
		"john",
		"jane",
	})

	// Use the convenience function for slice payloads
	users, err := hop.PayloadAsSlice(event)
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

	// Create an event with a map payload
	event := hop.NewEvent("regions.updated", map[string]any{
		"us-east": Region{Code: "USE", Name: "US East"},
		"us-west": Region{Code: "USW", Name: "US West"},
	})

	// Convert the payload to a map of Regions
	regions, err := hop.PayloadMapAs[Region](event)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Process the typed regions
	for key, region := range regions {
		fmt.Printf("Region %s: %s (%s)\n", key, region.Name, region.Code)
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
	event := hop.NewEvent("users.imported", []any{
		User{ID: "1", Name: "Alice"},
		User{ID: "2", Name: "Bob"},
	})

	// Convert the payload to a slice of Users
	users, err := hop.PayloadSliceAs[User](event)
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
