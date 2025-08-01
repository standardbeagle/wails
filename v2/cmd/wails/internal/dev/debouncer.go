package dev

import (
	"sync"
	"time"
)

// Debouncer prevents rapid repeated calls
type Debouncer struct {
	delay  time.Duration
	timers map[string]*time.Timer
	mu     sync.Mutex
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:  delay,
		timers: make(map[string]*time.Timer),
	}
}

// Debounce executes the function after the delay, canceling any previous call with the same key
func (d *Debouncer) Debounce(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer
	if timer, exists := d.timers[key]; exists {
		timer.Stop()
	}

	// Create new timer
	d.timers[key] = time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()

		fn()
	})
}

// Cancel cancels the pending execution for the given key
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, exists := d.timers[key]; exists {
		timer.Stop()
		delete(d.timers, key)
	}
}

// Clear cancels all pending executions
func (d *Debouncer) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, timer := range d.timers {
		timer.Stop()
		delete(d.timers, key)
	}
}

// HasPending returns true if there are pending executions
func (d *Debouncer) HasPending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.timers) > 0
}

// PendingKeys returns the keys of all pending executions
func (d *Debouncer) PendingKeys() []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	keys := make([]string, 0, len(d.timers))
	for key := range d.timers {
		keys = append(keys, key)
	}
	return keys
}