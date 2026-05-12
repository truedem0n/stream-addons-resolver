package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/truedem0n/playbridge-stream-resolver/config"
)

// handleGetRateLimits serves GET /api/config/rate-limits — returns all profiles.
func (s *Server) handleGetRateLimits(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	profiles := s.cfg.RateLimitProfiles
	if profiles == nil {
		profiles = map[string]config.RateLimitProfile{}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(profiles)
}

// handleUpdateRateLimits serves PUT /api/config/rate-limits — replaces the entire profile set.
// Body: { "torbox": { "per_minute": 8, "per_hour": 50, "max_concurrent": 3 }, ... }
// Profiles not in the body are removed; profiles present are upserted.
// Empty body or {} clears all profiles.
func (s *Server) handleUpdateRateLimits(w http.ResponseWriter, r *http.Request) {
	var incoming map[string]config.RateLimitProfile
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	for name, p := range incoming {
		if name == "" {
			http.Error(w, "profile name cannot be empty", http.StatusBadRequest)
			return
		}
		if p.PerMinute < 0 || p.PerHour < 0 || p.MaxConcurrent < 0 {
			http.Error(w, "profile "+name+": limits must be non-negative", http.StatusBadRequest)
			return
		}
	}

	s.mu.Lock()
	s.cfg.RateLimitProfiles = incoming
	s.applyRuntimeConfig()
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[rate-limits] failed to persist config: %v", err)
		http.Error(w, "profiles updated but config could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[rate-limits] %d profile(s) configured", len(incoming))
	for name, p := range incoming {
		log.Printf("[rate-limits]   %s — per_minute:%d per_hour:%d max_concurrent:%d",
			name, p.PerMinute, p.PerHour, p.MaxConcurrent)
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(incoming)
}
