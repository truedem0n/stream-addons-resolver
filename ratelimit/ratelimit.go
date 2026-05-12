// Package ratelimit implements per-profile rate limiting for outbound calls
// to debrid services. Multiple source addons can share a single profile so
// e.g. AIOStreams and Meteor both backed by the same Torbox account count
// against one shared bucket rather than two independent ones.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// ProfileConfig defines the limits for a named rate-limit profile.
// Zero values disable that limit dimension entirely.
type ProfileConfig struct {
	PerMinute     int
	PerHour       int
	MaxConcurrent int
}

// Limiter manages multiple named profiles. Calls referencing the same profile
// name share its bucket and concurrency semaphore.
type Limiter struct {
	mu       sync.Mutex
	profiles map[string]*profileState
}

// New returns an empty Limiter. Register profiles with SetProfile before use.
func New() *Limiter {
	return &Limiter{profiles: make(map[string]*profileState)}
}

type profileState struct {
	cfg    ProfileConfig
	mu     sync.Mutex
	events []time.Time // sorted ascending; trimmed to last hour
	sem    chan struct{}
}

func newProfileState(cfg ProfileConfig) *profileState {
	p := &profileState{cfg: cfg}
	if cfg.MaxConcurrent > 0 {
		p.sem = make(chan struct{}, cfg.MaxConcurrent)
	}
	return p
}

// SetProfile installs or replaces a profile by name.
// If a profile with the same MaxConcurrent already exists, its bucket and
// semaphore are preserved (so in-flight calls aren't disrupted) and only the
// rate fields are updated. A MaxConcurrent change rebuilds the profile — any
// in-flight calls on the old semaphore continue to drain naturally.
func (l *Limiter) SetProfile(name string, cfg ProfileConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if existing, ok := l.profiles[name]; ok && existing.cfg.MaxConcurrent == cfg.MaxConcurrent {
		existing.mu.Lock()
		existing.cfg = cfg
		existing.mu.Unlock()
		return
	}
	l.profiles[name] = newProfileState(cfg)
}

// RemoveProfile deletes a profile. In-flight calls referencing it continue.
func (l *Limiter) RemoveProfile(name string) {
	l.mu.Lock()
	delete(l.profiles, name)
	l.mu.Unlock()
}

// SyncProfiles replaces the entire profile set with the provided map.
// Profiles not in the map are removed; profiles present are upserted.
func (l *Limiter) SyncProfiles(next map[string]ProfileConfig) {
	l.mu.Lock()
	for name := range l.profiles {
		if _, ok := next[name]; !ok {
			delete(l.profiles, name)
		}
	}
	l.mu.Unlock()

	for name, cfg := range next {
		l.SetProfile(name, cfg)
	}
}

func (l *Limiter) get(name string) *profileState {
	l.mu.Lock()
	p := l.profiles[name]
	l.mu.Unlock()
	return p
}

// Acquire reserves a slot for a probe call against profile.
// It blocks on the concurrency semaphore (until ctx is done) and then atomically
// reserves a token in the per-minute and per-hour windows.
//
// Returns ok=true with a release fn that must be called when the probe ends.
// Returns ok=false (with reason) when:
//   - the context expires while waiting for a semaphore slot
//   - the per-minute or per-hour bucket is exhausted
//
// An empty name or unknown profile returns ok=true with a no-op release.
// Tokens are NOT refunded on failure — the API call counts at the debrid
// service regardless of whether the response was useful to us.
func (l *Limiter) Acquire(ctx context.Context, name string) (release func(), ok bool, reason string) {
	if name == "" {
		return func() {}, true, ""
	}
	p := l.get(name)
	if p == nil {
		return func() {}, true, ""
	}

	if p.sem != nil {
		select {
		case p.sem <- struct{}{}:
		case <-ctx.Done():
			return nil, false, "context cancelled waiting for semaphore"
		}
	}

	if !p.tryReserve(time.Now()) {
		if p.sem != nil {
			<-p.sem
		}
		return nil, false, "bucket exhausted"
	}

	return func() {
		if p.sem != nil {
			<-p.sem
		}
	}, true, ""
}

// tryReserve checks both windows and commits an event under a single lock.
func (p *profileState) tryReserve(now time.Time) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	hourCutoff := now.Add(-time.Hour)
	i := 0
	for i < len(p.events) && p.events[i].Before(hourCutoff) {
		i++
	}
	if i > 0 {
		p.events = p.events[i:]
	}

	if p.cfg.PerMinute > 0 {
		minCutoff := now.Add(-time.Minute)
		count := 0
		for j := len(p.events) - 1; j >= 0; j-- {
			if p.events[j].After(minCutoff) {
				count++
			} else {
				break
			}
		}
		if count >= p.cfg.PerMinute {
			return false
		}
	}
	if p.cfg.PerHour > 0 && len(p.events) >= p.cfg.PerHour {
		return false
	}

	p.events = append(p.events, now)
	return true
}

// Stats reports current usage for a profile (for diagnostics / future UI).
// Returns zero values if the profile does not exist.
func (l *Limiter) Stats(name string) (perMinUsed, perHourUsed, concurrentUsed int) {
	p := l.get(name)
	if p == nil {
		return 0, 0, 0
	}
	now := time.Now()
	p.mu.Lock()
	hourCutoff := now.Add(-time.Hour)
	minCutoff := now.Add(-time.Minute)
	for _, t := range p.events {
		if t.After(hourCutoff) {
			perHourUsed++
		}
		if t.After(minCutoff) {
			perMinUsed++
		}
	}
	p.mu.Unlock()
	if p.sem != nil {
		concurrentUsed = len(p.sem)
	}
	return
}
