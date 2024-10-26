package check

// Numeric represents types that can be compared numerically
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	~float32 | ~float64
}

// ValidationFunc is a non-generic function type for validation
type ValidationFunc func(value any) bool
