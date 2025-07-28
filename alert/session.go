package alert

import "context"

// SessionManager defines the interface for managing session data.
type SessionManager interface {
	Get(ctx context.Context, key string) interface{}
	Pop(ctx context.Context, key string) interface{}
	Put(ctx context.Context, key string, val interface{})
}
