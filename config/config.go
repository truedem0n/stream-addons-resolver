package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SourceAddon is a configured upstream Stremio stream addon.
type SourceAddon struct {
	URL              string `json:"url"`
	Name             string `json:"name"`
	Priority         int    `json:"priority"`             // lower = higher priority in merge
	TimeoutMs        int    `json:"timeout_ms"`           // per-addon HTTP timeout, default 8000
	RateLimitProfile string `json:"rate_limit_profile,omitempty"` // profile name from RateLimitProfiles; empty = no limit
	SkipProbe        bool   `json:"skip_probe,omitempty"` // true → addon's streams are accepted without probe validation
}

// MetaAddon is a Stremio meta addon used to look up runtime for probing validation.
// Addons are queried in order; the first one that returns a runtime wins.
type MetaAddon struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	TimeoutMs int    `json:"timeout_ms"` // per-request HTTP timeout, default 5000
}

// ProbeCacheConfig controls TTLs for the in-memory probe outcome cache.
// Failures get a shorter TTL than successes so transient debrid issues can recover.
type ProbeCacheConfig struct {
	SuccessTTLSeconds int `json:"success_ttl_seconds"` // default 3600
	FailureTTLSeconds int `json:"failure_ttl_seconds"` // default 600
}

// ProbingConfig controls ffprobe-based stream validation.
type ProbingConfig struct {
	Enabled       bool             `json:"enabled"`
	MaxAttempts   int              `json:"max_attempts"`               // legacy field, kept for config compatibility
	MaxCandidates int              `json:"max_candidates"`             // top-N candidates to consider per pass, default 5
	TimeoutMs     int              `json:"timeout_ms"`                 // per-probe ffprobe timeout, default 15000
	EarlyExit     *bool            `json:"early_exit,omitempty"`       // return as soon as best passer confirmed; default true
	Cache         ProbeCacheConfig `json:"cache"`                      // probe outcome cache TTLs
}

// EarlyExitEnabled reports whether early-exit probing is on.
// Defaults to true when the field is absent from config.
func (p ProbingConfig) EarlyExitEnabled() bool {
	return p.EarlyExit == nil || *p.EarlyExit
}

// RateLimitProfile defines limits shared by all addons referencing it by name.
// Zero values disable that limit dimension. Multiple addons backed by the same
// debrid service should share one profile so their probe traffic is bucketed
// together rather than each getting an independent allowance.
type RateLimitProfile struct {
	PerMinute     int `json:"per_minute"`
	PerHour       int `json:"per_hour"`
	MaxConcurrent int `json:"max_concurrent"`
}

// CacheConfig controls TTLs for on-disk caches.
type CacheConfig struct {
	StreamListTTLSeconds int `json:"stream_list_ttl_seconds"` // merged stream list, default 300
	PlayResultTTLSeconds int `json:"play_result_ttl_seconds"` // resolved play URL, default 3600
	MetaTTLSeconds       int `json:"meta_ttl_seconds"`        // runtime lookup cache, default 86400
}

// Bucket is a two-list preference set for a single filter dimension.
// Preferred streams are tried first; if all fail, everything not in Excluded
// is retried. Excluded streams are never used.
type Bucket struct {
	Preferred []string `json:"preferred"`
	Excluded  []string `json:"excluded"`
}

// UnmarshalJSON accepts the new object form `{preferred, excluded}` and is
// tolerant of legacy shapes:
//   - a JSON string is treated as a single preferred value (skipping ""/"auto")
//   - a JSON array of strings is treated as the preferred list
//   - null leaves the bucket zero-valued
//
// Unknown fields like the old "fallback" key are silently ignored.
func (b *Bucket) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	switch trimmed[0] {
	case '{':
		type bucketAlias Bucket
		var alias bucketAlias
		if err := json.Unmarshal(data, &alias); err != nil {
			return err
		}
		*b = Bucket(alias)
		return nil
	case '[':
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		b.Preferred = arr
		return nil
	case '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s != "" && !strings.EqualFold(s, "auto") {
			b.Preferred = []string{s}
		}
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into Bucket", string(trimmed))
}

// DefaultsConfig holds server-side streaming preference defaults applied when
// a /play request omits the corresponding query parameter.
// Explicit query params always take precedence over these defaults.
type DefaultsConfig struct {
	Quality      Bucket   `json:"quality"`
	SourceType   Bucket   `json:"source_type"`
	Source       Bucket   `json:"source"`
	AudioLang    Bucket   `json:"audio_lang"`
	ExcludeWords []string `json:"exclude_words"` // case-insensitive substrings — streams matching any are dropped
	MinSize      float64  `json:"min_size"`      // GB — 0 = no limit
	MaxSize      float64  `json:"max_size"`      // GB — 0 = no limit
	MinBitrate   float64  `json:"min_bitrate"`   // Mbps — 0 = no limit
	MaxBitrate   float64  `json:"max_bitrate"`   // Mbps — 0 = no limit
}

// Config is the top-level addon configuration.
type Config struct {
	Port              int                         `json:"port"`     // HTTP listen port, default 7000
	BaseURL           string                      `json:"base_url"` // public base URL e.g. http://localhost:7000
	Addons            []SourceAddon               `json:"addons"`
	MetaAddons        []MetaAddon                 `json:"meta_addons"` // used to fetch runtime for probing validation
	Probing           ProbingConfig               `json:"probing"`
	Cache             CacheConfig                 `json:"cache"`
	Defaults          DefaultsConfig              `json:"defaults"`
	RateLimitProfiles map[string]RateLimitProfile `json:"rate_limit_profiles,omitempty"`
}

// Load reads and parses the config file at the given path, applying defaults
// for any missing numeric fields.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	ApplyDefaults(&cfg)
	return &cfg, nil
}

// ApplyDefaults fills in any zero-valued numeric fields with sensible defaults.
// Safe to call multiple times — only zero values are touched.
func ApplyDefaults(cfg *Config) {
	if cfg.Port == 0 {
		cfg.Port = 7000
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	}

	for i := range cfg.Addons {
		if cfg.Addons[i].TimeoutMs == 0 {
			cfg.Addons[i].TimeoutMs = 8000
		}
		if cfg.Addons[i].Name == "" {
			cfg.Addons[i].Name = fmt.Sprintf("Addon %d", i+1)
		}
	}

	for i := range cfg.MetaAddons {
		if cfg.MetaAddons[i].TimeoutMs == 0 {
			cfg.MetaAddons[i].TimeoutMs = 5000
		}
		if cfg.MetaAddons[i].Name == "" {
			cfg.MetaAddons[i].Name = fmt.Sprintf("Meta Addon %d", i+1)
		}
	}

	if cfg.Probing.MaxAttempts == 0 {
		cfg.Probing.MaxAttempts = 5
	}
	if cfg.Probing.MaxCandidates == 0 {
		cfg.Probing.MaxCandidates = 5
	}
	if cfg.Probing.TimeoutMs == 0 {
		cfg.Probing.TimeoutMs = 15000
	}
	if cfg.Probing.Cache.SuccessTTLSeconds == 0 {
		cfg.Probing.Cache.SuccessTTLSeconds = 3600
	}
	if cfg.Probing.Cache.FailureTTLSeconds == 0 {
		cfg.Probing.Cache.FailureTTLSeconds = 600
	}

	if cfg.Cache.StreamListTTLSeconds == 0 {
		cfg.Cache.StreamListTTLSeconds = 300
	}
	if cfg.Cache.PlayResultTTLSeconds == 0 {
		cfg.Cache.PlayResultTTLSeconds = 3600
	}
	if cfg.Cache.MetaTTLSeconds == 0 {
		cfg.Cache.MetaTTLSeconds = 86400
	}
}
