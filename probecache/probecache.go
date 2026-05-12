// Package probecache holds an in-memory cache of probe outcomes keyed by URL.
// Successful and failed probes carry separate TTLs so persistently-bad URLs
// (debrid links that never expose duration) stop wasting rate-limit budget on
// repeated probes, while still re-checking after a short recovery window.
package probecache

import (
	"sync"
	"time"
)

// Outcome records a single probe attempt.
// DurationMins is 0 when the probe failed; Err carries the failure reason.
type Outcome struct {
	DurationMins int
	Err          string
	ProbedAt     time.Time
}

// Failed reports whether this outcome represents a failure (error or zero duration).
func (o Outcome) Failed() bool {
	return o.Err != "" || o.DurationMins == 0
}

// Cache is a thread-safe URL → Outcome store with split success/failure TTLs.
type Cache struct {
	mu         sync.RWMutex
	entries    map[string]entry
	successTTL time.Duration
	failureTTL time.Duration
}

type entry struct {
	outcome   Outcome
	expiresAt time.Time
}

// New returns a Cache and starts a background goroutine that periodically
// removes expired entries.
func New(successTTL, failureTTL time.Duration) *Cache {
	c := &Cache{
		entries:    make(map[string]entry),
		successTTL: successTTL,
		failureTTL: failureTTL,
	}
	go c.gc()
	return c
}

// SetTTLs updates the TTLs used for newly-stored outcomes.
// Existing entries keep their original expiration.
func (c *Cache) SetTTLs(successTTL, failureTTL time.Duration) {
	c.mu.Lock()
	c.successTTL = successTTL
	c.failureTTL = failureTTL
	c.mu.Unlock()
}

// Get returns the cached outcome for url, or false if no fresh entry exists.
func (c *Cache) Get(url string) (Outcome, bool) {
	c.mu.RLock()
	e, ok := c.entries[url]
	c.mu.RUnlock()
	if !ok {
		return Outcome{}, false
	}
	if time.Now().After(e.expiresAt) {
		c.mu.Lock()
		delete(c.entries, url)
		c.mu.Unlock()
		return Outcome{}, false
	}
	return e.outcome, true
}

// Put stores an outcome with the appropriate TTL for its success/failure state.
func (c *Cache) Put(url string, o Outcome) {
	c.mu.Lock()
	ttl := c.successTTL
	if o.Failed() {
		ttl = c.failureTTL
	}
	c.entries[url] = entry{
		outcome:   o,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Len returns the current entry count (useful for diagnostics).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) gc() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
