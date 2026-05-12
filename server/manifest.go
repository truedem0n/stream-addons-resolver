package server

import (
	"encoding/json"
	"net/http"

	"github.com/truedem0n/playbridge-stream-resolver/types"
	"github.com/truedem0n/playbridge-stream-resolver/version"
)

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	manifest := types.Manifest{
		ID:          "com.playbridge.stream-resolver",
		Name:        "Stream Resolver",
		Description: "Aggregates streams from multiple Stremio addons, ranks them, and validates with ffprobe.",
		Version:     version.Version,
		Resources:   []string{"stream"},
		Types:       []string{"movie", "series"},
		Catalogs:    []any{},
		BehaviorHints: types.ManifestHints{
			Configurable: true,
			PlayEndpoint: "/api/play/{type}/{id}",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)
}
