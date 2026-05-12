package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/truedem0n/playbridge-stream-resolver/config"
)

// handleGetFullConfig serves GET /api/config — returns the full config as JSON.
func (s *Server) handleGetFullConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(cfg)
}

// handleUpdateFullConfig serves PUT /api/config — replaces the entire config.
// Applies defaults and re-syncs runtime services (probe cache, rate limiter)
// so a raw edit takes effect without restarting the server.
func (s *Server) handleUpdateFullConfig(w http.ResponseWriter, r *http.Request) {
	var incoming config.Config
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	config.ApplyDefaults(&incoming)

	s.mu.Lock()
	s.cfg = &incoming
	s.applyRuntimeConfig()
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[config] failed to persist full config: %v", err)
		http.Error(w, "config updated in memory but could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[config] full config replaced and saved")
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(s.cfg)
}
