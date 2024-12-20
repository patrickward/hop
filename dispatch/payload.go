package dispatch

import (
	"context"
	"fmt"
)

// PayloadAs safely converts an event's payload to the specified type T.
// Returns the typed payload and any conversion error.
func PayloadAs[T any](e Event) (T, error) {
	var zero T
	if e.Payload == nil {
		return zero, fmt.Errorf("event payload is nil")
	}

	payload, ok := e.Payload.(T)
	if !ok {
		return zero, fmt.Errorf("invalid payload type: expected %T, got %T", zero, e.Payload)
	}

	return payload, nil
}

// MustPayloadAs converts an event's payload to the specified type T.
// Panics if the conversion fails.
func MustPayloadAs[T any](e Event) T {
	payload, err := PayloadAs[T](e)
	if err != nil {
		panic(err)
	}
	return payload
}

// HandlePayload creates an event handler that automatically converts the payload
// to the specified type T and calls the provided typed handler function.
// If type conversion fails, logs the error and returns without calling the handler.
func HandlePayload[T any](handler func(context.Context, T)) Handler {
	return func(ctx context.Context, e Event) {
		payload, err := PayloadAs[T](e)
		if err != nil {
			return
		}
		handler(ctx, payload)
	}
}

// IsPayloadType checks if an event's payload is of the specified type T.
func IsPayloadType[T any](e Event) bool {
	_, ok := e.Payload.(T)
	return ok
}

// PayloadAsMap is a convenience function for working with map[string]any payloads,
// which are common when dealing with JSON data.
func PayloadAsMap(e Event) (map[string]any, error) {
	return PayloadAs[map[string]any](e)
}

// PayloadMapAs converts a map payload into a map with typed values.
// Returns an error if the payload is not a map or if any value cannot be converted to type T.
//
// Example:
//
//	type User struct { ID string }
//	userMap, err := PayloadMapAs[User](event)
func PayloadMapAs[T any](e Event) (map[string]T, error) {
	rawMap, err := PayloadAsMap(e)
	if err != nil {
		return nil, fmt.Errorf("payload is not a map: %w", err)
	}

	result := make(map[string]T, len(rawMap))
	for key, val := range rawMap {
		if val == nil {
			continue
		}
		typed, ok := val.(T)
		if !ok {
			return nil, fmt.Errorf("invalid type for key %q: expected %T, got %T", key, *new(T), val)
		}
		result[key] = typed
	}

	return result, nil
}

// PayloadAsSlice is a convenience function for working with []any payloads.
func PayloadAsSlice(e Event) ([]any, error) {
	return PayloadAs[[]any](e)
}

// PayloadSliceAs converts a slice payload into a slice of typed elements.
// Returns an error if the payload is not a slice or if any element cannot be converted to type T.
//
// Example:
//
//	type User struct { ID string }
//	users, err := PayloadSliceAs[User](event)
func PayloadSliceAs[T any](e Event) ([]T, error) {
	slice, err := PayloadAsSlice(e)
	if err != nil {
		return nil, fmt.Errorf("payload is not a slice: %w", err)
	}

	result := make([]T, 0, len(slice))
	for i, item := range slice {
		if item == nil {
			continue
		}
		typed, ok := item.(T)
		if !ok {
			return nil, fmt.Errorf("invalid type at index %d: expected %T, got %T", i, *new(T), item)
		}
		result = append(result, typed)
	}

	return result, nil
}
