package route

import (
	"net/http"
)

// Middleware represents a function that wraps an http.Handler with additional functionality
type Middleware func(http.Handler) http.Handler

// Chain represents an immutable chain of http.Handler middleware
//
// This is essentially the same as the `github.com/justinas/alice` package,
// which allows for chaining middleware in a clean and reusable way. All credit goes to `github.com/justinas`.
//
// I reimplemented it with a new name to match my needs, and added some additional functionality to it. All credit goes to `github.com/justinas`.
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain, memoizing the middlewares
func NewChain(middleware ...Middleware) Chain {
	return Chain{append(([]Middleware)(nil), middleware...)}
}

// Extend adds a chain by adding the provided chain's middleware to the current chain
// 1. It returns a new chain with the combined middlewares
// 2. The original chain remains unchanged
// 3. This allows for composing middleware chains easily
//
// Example:
// chain1 := wrap.New(middleware1, middleware2)
// chain2 := wrap.New(middleware3)
// combinedChain := chain1.Extend(chain2)

func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.middlewares...)
}

// Append adds additional middleware to the chain and returns a new chain
// 1. It returns a new chain with the combined middlewares
// 2. The original chain remains unchanged
// 3. This allows for composing middleware chains easily
//
// Example:
// chain := wrap.New(middleware1).Append(middleware2, middleware3)
func (c Chain) Append(middleware ...Middleware) Chain {
	newMid := make([]Middleware, 0, len(c.middlewares)+len(middleware))
	newMid = append(newMid, c.middlewares...)
	newMid = append(newMid, middleware...)
	return Chain{middlewares: newMid}
}

// Then chains the middleware to the given http.Handler
// and returns the resulting http.Handler Chains can be safely reused,
// and the original chain remains unchanged.
//
//	Example:
//
// chain := wrap.New(middleware1, middleware2)
// pipe1 := chain.Then(anotherHandler)
// pipe2 := chain.Then(yetAnotherHandler)
func (c Chain) Then(h http.Handler) http.Handler {
	return Around(h, c.middlewares...)
}

// ThenFunc wraps the given handler function with all middleware in the chain
// and returns the resulting http.Handler. If the provided handler function is nil,
// it returns a handler that does nothing.
//
// Example:
// chain := wrap.New(middleware1, middleware2)
// pipe1 := chain.ThenFunc(myHandlerFunc)
// pipe2 : = chain.ThenFunc(anotherHandlerFunc)
func (c Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	// This nil check cannot be removed due to the "nil is not nil" common mistake in Go.
	// Required due to: https://stackoverflow.com/questions/33426977/how-to-golang-check-a-variable-is-nil
	if fn == nil {
		return c.Then(nil)
	}
	return c.Then(fn)
}

// Around wraps a handler with the provided middleware in the order they are passed
// It return the resulting http.Handler. So, it's mostly useful for on-the-fly middleware application.
//
// Example:
// h := wrap.Around(myHandler, middleware1, middleware2) // h is now wrapped with middleware1 and middleware2
func Around(h http.Handler, middleware ...Middleware) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// Before is an alias for Around to provide a more semantic API
func Before(h http.Handler, middleware ...Middleware) http.Handler {
	return Around(h, middleware...)
}

// After wraps a handler with the provided middleware in reverse order.
// It returns the resulting http.Handler. This is useful for applying middleware that should run after the main handler
//
// Example:
// h := wrap.After(myHandler, middleware1, middleware2) // h is now wrapped with middleware2 and middleware1, reversed
func After(h http.Handler, middleware ...Middleware) http.Handler {
	for i := 0; i < len(middleware); i++ {
		h = middleware[i](h)
	}
	return h
}
