package bloom

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

// Filter is a thread-safe in-process bloom filter.
type Filter struct {
	mu sync.RWMutex
	bf *bloom.BloomFilter
}

// New creates a filter sized for n expected entries at the given false-positive rate.
func New(n uint, fpRate float64) *Filter {
	return &Filter{bf: bloom.NewWithEstimates(n, fpRate)}
}

// Add inserts item into the filter.
func (f *Filter) Add(item string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bf.AddString(item)
}

// Contains reports whether item is possibly in the filter.
// A false return is definitive (not present); a true return may be a false positive.
func (f *Filter) Contains(item string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.bf.TestString(item)
}
