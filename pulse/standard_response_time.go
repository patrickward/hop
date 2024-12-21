package pulse

import (
	"math"
	"sort"
	"sync"
)

// ResponseTimeTracker keeps track of response time metrics
type responseTimeTracker struct {
	samples    []float64
	currentIdx int
	size       int
	sum        float64
	count      int64
	mu         sync.RWMutex
}

func newResponseTimeTracker(size int) *responseTimeTracker {
	return &responseTimeTracker{
		samples: make([]float64, size),
		size:    size,
	}
}

func (rt *responseTimeTracker) Record(duration float64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// Update circular buffer
	rt.samples[rt.currentIdx] = duration
	rt.currentIdx = (rt.currentIdx + 1) % rt.size

	// Update running statistics
	rt.sum += duration
	rt.count++
}

func (rt *responseTimeTracker) GetPercentile(p float64) float64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.count == 0 {
		return 0
	}

	// Create copy of valid samples
	numSamples := int(math.Min(float64(rt.count), float64(rt.size)))
	samples := make([]float64, numSamples)
	copy(samples, rt.samples[:numSamples])
	sort.Float64s(samples)

	// Calculate percentile index
	idx := int(float64(numSamples-1) * p / 100)
	return samples[idx]
}

func (rt *responseTimeTracker) GetAverage() float64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.count == 0 {
		return 0
	}
	return rt.sum / float64(rt.count)
}
