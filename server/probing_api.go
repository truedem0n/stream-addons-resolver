package server

import (
	"encoding/json"
	"log"
	"net/http"
)

// probingConfigResponse is the shape returned/accepted by the probing config API.
type probingConfigResponse struct {
	Enabled           bool `json:"enabled"`
	MaxAttempts       int  `json:"max_attempts"`
	MaxCandidates     int  `json:"max_candidates"`
	TimeoutMs         int  `json:"timeout_ms"`
	EarlyExit         bool `json:"early_exit"`
	SuccessTTLSeconds int  `json:"success_ttl_seconds"`
	FailureTTLSeconds int  `json:"failure_ttl_seconds"`
}

// handleGetProbingConfig serves GET /api/config/probing
func (s *Server) handleGetProbingConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	resp := probingConfigResponse{
		Enabled:           s.cfg.Probing.Enabled,
		MaxAttempts:       s.cfg.Probing.MaxAttempts,
		MaxCandidates:     s.cfg.Probing.MaxCandidates,
		TimeoutMs:         s.cfg.Probing.TimeoutMs,
		EarlyExit:         s.cfg.Probing.EarlyExitEnabled(),
		SuccessTTLSeconds: s.cfg.Probing.Cache.SuccessTTLSeconds,
		FailureTTLSeconds: s.cfg.Probing.Cache.FailureTTLSeconds,
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleUpdateProbingConfig serves PUT /api/config/probing.
// All fields are optional — only fields present in the JSON body are applied.
func (s *Server) handleUpdateProbingConfig(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()

	if v, ok := raw["enabled"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid value for enabled: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Probing.Enabled = b
	}

	if v, ok := raw["max_attempts"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n <= 0 {
			s.mu.Unlock()
			http.Error(w, "max_attempts must be a positive integer", http.StatusBadRequest)
			return
		}
		s.cfg.Probing.MaxAttempts = n
	}

	if v, ok := raw["max_candidates"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n <= 0 {
			s.mu.Unlock()
			http.Error(w, "max_candidates must be a positive integer", http.StatusBadRequest)
			return
		}
		s.cfg.Probing.MaxCandidates = n
	}

	if v, ok := raw["timeout_ms"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n <= 0 {
			s.mu.Unlock()
			http.Error(w, "timeout_ms must be a positive integer", http.StatusBadRequest)
			return
		}
		s.cfg.Probing.TimeoutMs = n
	}

	if v, ok := raw["early_exit"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			s.mu.Unlock()
			http.Error(w, "invalid value for early_exit: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Probing.EarlyExit = &b
	}

	if v, ok := raw["success_ttl_seconds"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n < 0 {
			s.mu.Unlock()
			http.Error(w, "success_ttl_seconds must be a non-negative integer", http.StatusBadRequest)
			return
		}
		s.cfg.Probing.Cache.SuccessTTLSeconds = n
	}

	if v, ok := raw["failure_ttl_seconds"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil || n < 0 {
			s.mu.Unlock()
			http.Error(w, "failure_ttl_seconds must be a non-negative integer", http.StatusBadRequest)
			return
		}
		s.cfg.Probing.Cache.FailureTTLSeconds = n
	}

	s.applyRuntimeConfig()

	resp := probingConfigResponse{
		Enabled:           s.cfg.Probing.Enabled,
		MaxAttempts:       s.cfg.Probing.MaxAttempts,
		MaxCandidates:     s.cfg.Probing.MaxCandidates,
		TimeoutMs:         s.cfg.Probing.TimeoutMs,
		EarlyExit:         s.cfg.Probing.EarlyExitEnabled(),
		SuccessTTLSeconds: s.cfg.Probing.Cache.SuccessTTLSeconds,
		FailureTTLSeconds: s.cfg.Probing.Cache.FailureTTLSeconds,
	}
	err := s.persistConfig()
	s.mu.Unlock()

	if err != nil {
		log.Printf("[probing] failed to persist config: %v", err)
		http.Error(w, "settings updated but config could not be saved: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[probing] config updated — enabled:%v maxCandidates:%d timeoutMs:%d earlyExit:%v cache(success:%ds fail:%ds)",
		resp.Enabled, resp.MaxCandidates, resp.TimeoutMs, resp.EarlyExit,
		resp.SuccessTTLSeconds, resp.FailureTTLSeconds)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
