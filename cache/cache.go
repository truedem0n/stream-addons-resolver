package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store is a simple on-disk JSON cache with TTL-based expiry.
// Each entry is a JSON file under a named subdirectory of the cache root.
// The file's modification time is used as the write timestamp.
type Store struct {
	mu  sync.RWMutex
	dir string
	ttl time.Duration
}

// NewStore creates a Store rooted at dir/namespace with the given TTL.
func NewStore(root, namespace string, ttl time.Duration) *Store {
	dir := filepath.Join(root, namespace)
	os.MkdirAll(dir, 0755)
	return &Store{dir: dir, ttl: ttl}
}

// Get retrieves a cached value into dest. Returns (false, nil) on miss or expiry.
func (s *Store) Get(key string, dest any) (bool, error) {
	s.mu.RLock()
	ttl := s.ttl
	s.mu.RUnlock()

	path := s.path(key)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if time.Since(info.ModTime()) > ttl {
		return false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

// Set writes value to cache under key.
func (s *Store) Set(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(key), data, 0644)
}

// SetTTL updates the TTL for future reads. Thread-safe.
func (s *Store) SetTTL(ttl time.Duration) {
	s.mu.Lock()
	s.ttl = ttl
	s.mu.Unlock()
}

// Clear removes all entries from this cache store.
func (s *Store) Clear() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var errs []error
	for _, e := range entries {
		if err := os.Remove(filepath.Join(s.dir, e.Name())); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("cleared with %d errors; first: %w", len(errs), errs[0])
	}
	return nil
}

func (s *Store) path(key string) string {
	// Sanitize key to a safe filename.
	safe := fmt.Sprintf("%x", []byte(key))
	return filepath.Join(s.dir, safe+".json")
}

// InflightMap prevents duplicate concurrent work for the same key.
// Call Start to claim a key; it returns true if the caller won the race.
// The caller must call Done when finished, which unblocks any waiters.
type InflightMap struct {
	mu      sync.Mutex
	entries map[string]*inflightEntry
}

type inflightEntry struct {
	done chan struct{}
}

func NewInflightMap() *InflightMap {
	return &InflightMap{entries: make(map[string]*inflightEntry)}
}

// Start claims key for the caller. Returns (true, noop) if the caller is first.
// Returns (false, waitFn) if another goroutine is already working; the caller
// should call waitFn() to block until the work is done, then read from cache.
func (m *InflightMap) Start(key string) (first bool, wait func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.entries[key]; ok {
		ch := e.done
		return false, func() { <-ch }
	}

	e := &inflightEntry{done: make(chan struct{})}
	m.entries[key] = e
	return true, nil
}

// Done releases key and unblocks any waiters.
func (m *InflightMap) Done(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.entries[key]; ok {
		close(e.done)
		delete(m.entries, key)
	}
}
