package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/truedem0n/playbridge-stream-resolver/config"
)

// handleGetAddons serves GET /api/addons
// Returns the current addon list as JSON.
func (s *Server) handleGetAddons(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	addons := s.cfg.Addons
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addons)
}

// handleAddAddon serves POST /api/addons
// Body: { "url": "...", "name": "...", "priority": 0, "timeout_ms": 8000 }
// Appends the addon to the live config and persists to disk.
func (s *Server) handleAddAddon(w http.ResponseWriter, r *http.Request) {
	var a config.SourceAddon
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	a.URL = strings.TrimSpace(a.URL)
	if a.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	if a.TimeoutMs == 0 {
		a.TimeoutMs = 8000
	}

	// Verify the manifest is reachable and valid before accepting the addon.
	manifestName, err := fetchManifest(a.URL, a.TimeoutMs)
	if err != nil {
		http.Error(w, "manifest check failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	if a.Name == "" {
		a.Name = manifestName
	}

	s.mu.Lock()
	// Reject duplicates by URL.
	for _, existing := range s.cfg.Addons {
		if existing.URL == a.URL {
			s.mu.Unlock()
			http.Error(w, "addon with this URL already exists", http.StatusConflict)
			return
		}
	}
	// Auto-assign priority = last position if not set.
	if a.Priority == 0 && len(s.cfg.Addons) > 0 {
		a.Priority = s.cfg.Addons[len(s.cfg.Addons)-1].Priority + 1
	}
	s.cfg.Addons = append(s.cfg.Addons, a)
	persistErr := s.persistConfig()
	s.mu.Unlock()

	if persistErr != nil {
		log.Printf("[addons] failed to persist config: %v", persistErr)
		http.Error(w, "addon added but config could not be saved: "+persistErr.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[addons] added %s (%s)", a.Name, a.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// handleDeleteAddon serves DELETE /api/addons?url=<encoded-url>
// Removes the addon with the matching URL from the live config and persists.
func (s *Server) handleDeleteAddon(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "url query param is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	before := len(s.cfg.Addons)
	filtered := s.cfg.Addons[:0]
	for _, a := range s.cfg.Addons {
		if a.URL != targetURL {
			filtered = append(filtered, a)
		}
	}
	if len(filtered) == before {
		s.mu.Unlock()
		http.Error(w, "addon not found", http.StatusNotFound)
		return
	}
	s.cfg.Addons = filtered
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[addons] failed to persist config: %v", err)
		http.Error(w, "addon removed but config could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[addons] removed %s", targetURL)
	w.WriteHeader(http.StatusNoContent)
}

// handlePatchAddon serves PATCH /api/addons?url=<encoded-url>
// Body: { "rate_limit_profile": "...", "skip_probe": true, "timeout_ms": 8000, "name": "...", "priority": 0 }
// All fields are optional — only fields present in the JSON body are applied.
func (s *Server) handlePatchAddon(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "url query param is required", http.StatusBadRequest)
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	idx := -1
	for i, a := range s.cfg.Addons {
		if a.URL == targetURL {
			idx = i
			break
		}
	}
	if idx < 0 {
		s.mu.Unlock()
		http.Error(w, "addon not found", http.StatusNotFound)
		return
	}

	if v, ok := raw["rate_limit_profile"]; ok {
		var p string
		if err := json.Unmarshal(v, &p); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid rate_limit_profile: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Addons[idx].RateLimitProfile = p
	}
	if v, ok := raw["skip_probe"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid skip_probe: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Addons[idx].SkipProbe = b
	}
	if v, ok := raw["timeout_ms"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n <= 0 {
			s.mu.Unlock()
			http.Error(w, "timeout_ms must be a positive integer", http.StatusBadRequest)
			return
		}
		s.cfg.Addons[idx].TimeoutMs = n
	}
	if v, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid name: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Addons[idx].Name = name
	}
	if v, ok := raw["disabled"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid disabled: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Addons[idx].Disabled = b
	}

	updated := s.cfg.Addons[idx]
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[addons] failed to persist patch: %v", err)
		http.Error(w, "addon updated but config could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[addons] patched %s — profile:%q skip_probe:%v disabled:%v",
		updated.Name, updated.RateLimitProfile, updated.SkipProbe, updated.Disabled)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// handleRefreshAddon serves POST /api/addons/refresh?url=<encoded-url>
// Re-fetches the addon's manifest to verify reachability, updates the stored
// name from the live manifest, and returns the updated addon. Leaves Disabled
// state untouched — refresh is a probe, not a state change.
func (s *Server) handleRefreshAddon(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "url query param is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	var existing config.SourceAddon
	found := false
	for _, a := range s.cfg.Addons {
		if a.URL == targetURL {
			existing = a
			found = true
			break
		}
	}
	s.mu.RUnlock()

	if !found {
		http.Error(w, "addon not found", http.StatusNotFound)
		return
	}

	timeoutMs := existing.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 8000
	}
	name, err := fetchManifest(existing.URL, timeoutMs)
	if err != nil {
		log.Printf("[addons] refresh failed for %s: %v", existing.URL, err)
		http.Error(w, "manifest check failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	s.mu.Lock()
	idx := -1
	for i, a := range s.cfg.Addons {
		if a.URL == targetURL {
			idx = i
			break
		}
	}
	if idx < 0 {
		s.mu.Unlock()
		http.Error(w, "addon not found", http.StatusNotFound)
		return
	}
	if name != "" {
		s.cfg.Addons[idx].Name = name
	}
	updated := s.cfg.Addons[idx]
	persistErr := s.persistConfig()
	s.mu.Unlock()

	if persistErr != nil {
		log.Printf("[addons] refresh persist failed: %v", persistErr)
		http.Error(w, "refreshed but config could not be saved: "+persistErr.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[addons] refreshed %s (%s)", updated.Name, updated.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// handleReorderAddons serves PUT /api/addons
// Body: ["url1","url2",...] — full ordered list of addon URLs.
// Resets priorities to match the new position (0 = first).
func (s *Server) handleReorderAddons(w http.ResponseWriter, r *http.Request) {
	var urls []string
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	byURL := make(map[string]config.SourceAddon, len(s.cfg.Addons))
	for _, a := range s.cfg.Addons {
		byURL[a.URL] = a
	}
	reordered := make([]config.SourceAddon, 0, len(urls))
	for i, u := range urls {
		a, ok := byURL[u]
		if !ok {
			s.mu.Unlock()
			http.Error(w, "unknown addon URL: "+u, http.StatusBadRequest)
			return
		}
		a.Priority = i
		reordered = append(reordered, a)
	}
	s.cfg.Addons = reordered
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[addons] failed to persist reorder: %v", err)
		http.Error(w, "reordered but config could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[addons] reordered — %d addons", len(reordered))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reordered)
}

// persistConfig writes the current in-memory config back to disk.
// Must be called with s.mu held (write lock).
func (s *Server) persistConfig() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	return os.WriteFile(s.cfgPath, data, 0644)
}

// fetchManifest fetches the addon's manifest.json, validates it, and returns
// the addon name. Returns an error if the URL is unreachable, returns a
// non-200 status, or does not decode as a valid Stremio manifest.
func fetchManifest(addonURL string, timeoutMs int) (string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(addonURL, "/"), "/manifest.json")
	manifestURL := base + "/manifest.json"

	client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	resp, err := client.Get(manifestURL)
	if err != nil {
		return "", fmt.Errorf("could not reach %s: %w", manifestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned HTTP %d", manifestURL, resp.StatusCode)
	}

	var m struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("invalid manifest JSON from %s: %w", manifestURL, err)
	}
	if m.Name == "" && m.ID == "" {
		return "", fmt.Errorf("manifest from %s is missing both id and name fields", manifestURL)
	}

	name := m.Name
	if name == "" {
		name = addonNameFromURL(addonURL)
	}
	return name, nil
}

// addonNameFromURL derives a human-readable name from an addon URL.
// Used as a fallback when the manifest cannot be fetched.
func addonNameFromURL(u string) string {
	u = strings.TrimSuffix(u, "/manifest.json")
	u = strings.TrimSuffix(u, "/")
	parts := strings.Split(u, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return u
}
