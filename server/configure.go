package server

import (
	"fmt"
	"net/http"
)

// handleConfigure serves GET /configure
func (s *Server) handleConfigure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, configurePage())
}

func configurePage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Stream Resolver — Configure</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #0f0f0f; color: #e0e0e0;
      min-height: 100vh; padding: 2rem 1rem 4rem;
    }
    .container { max-width: 700px; margin: 0 auto; }

    /* ── Header ── */
    .page-header {
      display: flex; align-items: center; gap: 1rem;
      margin-bottom: 2rem; flex-wrap: wrap;
    }
    .page-header-text { flex: 1; }
    h1 { font-size: 1.4rem; font-weight: 700; color: #fff; }
    .subtitle { font-size: 0.8rem; color: #666; margin-top: 0.2rem; }
    .manifest-pill {
      font-size: 0.75rem; color: #777;
      background: #161616; border: 1px solid #252525;
      border-radius: 6px; padding: 0.3rem 0.6rem;
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
      max-width: 260px;
    }
    .btn-install {
      background: #7b2fbe; border: none; color: #fff;
      border-radius: 6px; padding: 0.45rem 1rem;
      font-size: 0.825rem; cursor: pointer; white-space: nowrap;
      text-decoration: none; flex-shrink: 0;
    }
    .btn-install:hover { background: #9b4fde; }

    /* ── Cards ── */
    .card {
      background: #151515; border: 1px solid #242424;
      border-radius: 12px; padding: 1.25rem; margin-bottom: 1.25rem;
    }
    h2 {
      font-size: 0.875rem; font-weight: 600; color: #aaa;
      text-transform: uppercase; letter-spacing: 0.05em;
      margin-bottom: 0.2rem;
    }
    .card-desc {
      font-size: 0.78rem; color: #555; margin-bottom: 1rem;
    }
    .section-label {
      font-size: 0.75rem; font-weight: 600; color: #666;
      text-transform: uppercase; letter-spacing: 0.04em;
      margin: 1.25rem 0 0.6rem;
    }
    .section-label:first-of-type { margin-top: 0.25rem; }

    /* ── Draggable addon rows ── */
    .dnd-list { display: flex; flex-direction: column; gap: 0.3rem; margin-bottom: 0.75rem; }
    .dnd-item {
      display: flex; align-items: center; gap: 0.6rem;
      background: #1c1c1c; border: 1px solid #2a2a2a; border-radius: 8px;
      padding: 0.5rem 0.6rem; cursor: grab; user-select: none;
    }
    .dnd-item.dragging  { opacity: 0.35; }
    .dnd-item.drag-over { border-color: #4caf50; background: #1a241a; }
    .dnd-handle { color: #444; font-size: 1rem; line-height: 1; flex-shrink: 0; }
    .dnd-info   { flex: 1; min-width: 0; }
    .dnd-name   { font-size: 0.875rem; font-weight: 500; color: #ddd;
                  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .dnd-url    { font-size: 0.72rem; color: #555;
                  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .badge {
      font-size: 0.68rem; background: #222; color: #666;
      border-radius: 4px; padding: 2px 6px; white-space: nowrap; flex-shrink: 0;
    }
    .btn-remove {
      background: none; border: 1px solid transparent; color: #5a2a2a;
      border-radius: 6px; padding: 3px 8px; font-size: 0.75rem;
      cursor: pointer; white-space: nowrap; flex-shrink: 0; transition: all 0.15s;
    }
    .btn-remove:hover { border-color: #6a2a2a; color: #c0392b; background: #1e1010; }
    .empty { color: #444; font-size: 0.825rem; padding: 0.4rem 0; }

    /* ── Add form ── */
    .add-form { border-top: 1px solid #1f1f1f; margin-top: 0.75rem; padding-top: 0.75rem; }
    .form-row { display: flex; gap: 0.5rem; flex-wrap: wrap; }
    input[type="text"], input[type="number"] {
      background: #0f0f0f; border: 1px solid #2a2a2a; color: #e0e0e0;
      border-radius: 6px; padding: 0.45rem 0.7rem; font-size: 0.825rem; outline: none;
    }
    input[type="text"]:focus, input[type="number"]:focus { border-color: #444; }
    .url-input     { flex: 1; min-width: 180px; }
    .name-input    { width: 130px; }
    .timeout-input { width: 100px; }
    .btn-add {
      background: #1a2e1a; border: 1px solid #2d5a2d; color: #4caf50;
      border-radius: 6px; padding: 0.45rem 1rem; font-size: 0.825rem;
      cursor: pointer; white-space: nowrap;
    }
    .btn-add:hover { background: #243e24; }

    /* ── Buttons ── */
    .btn-save {
      background: #152033; border: 1px solid #2a5a8a; color: #5b9bd5;
      border-radius: 6px; padding: 0.45rem 1.1rem; font-size: 0.825rem;
      cursor: pointer; white-space: nowrap;
    }
    .btn-save:hover { background: #1e3a5a; }
    .btn-clear {
      background: none; border: 1px solid #252525; color: #666;
      border-radius: 6px; padding: 0.35rem 0.8rem; font-size: 0.775rem;
      cursor: pointer; white-space: nowrap;
    }
    .btn-clear:hover { background: #1a1a1a; color: #aaa; }

    /* ── Status lines ── */
    .status { font-size: 0.775rem; margin-top: 0.5rem; min-height: 1.1em; }
    .ok  { color: #4caf50; }
    .err { color: #c0392b; }

    /* ── Bucket rows ── */
    .bucket-rows { display: flex; flex-direction: column; gap: 0.25rem; margin-bottom: 0.75rem; }
    .bucket-row {
      display: flex; align-items: center; justify-content: space-between;
      background: #1a1a1a; border: 1px solid #242424; border-radius: 8px;
      padding: 0.3rem 0.75rem;
    }
    .bucket-row-label { font-size: 0.85rem; color: #bbb; flex: 1; }
    .seg-group { display: flex; background: #111; border-radius: 5px; padding: 2px; gap: 2px; }
    .seg-btn {
      background: none; border: none; border-radius: 3px;
      padding: 0.18rem 0.55rem; font-size: 0.72rem;
      cursor: pointer; color: #444; transition: all 0.15s; white-space: nowrap;
    }
    .seg-btn:not(.active):hover { color: #999; background: #1c1c1c; }
    .seg-btn.active[data-val="preferred"] { background: #1a3a1a; color: #4caf50; }
    .seg-btn.active[data-val="fallback"]  { background: #1e1e1e; color: #666; }
    .seg-btn.active[data-val="excluded"]  { background: #3a1a1a; color: #c0392b; }
    .bucket-empty { font-size: 0.8rem; color: #444; padding: 0.3rem 0; }

    /* Filter grid */
    .filter-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
    .filter-field { display: flex; flex-direction: column; gap: 0.25rem; }
    .filter-field input[type="number"] { width: 100%; }
    .field-label { font-size: 0.75rem; color: #666; }

    /* ── Cache & Probing ── */
    .ttl-row { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.6rem; flex-wrap: wrap; }
    .ttl-label { width: 150px; font-size: 0.825rem; color: #bbb; flex-shrink: 0; }
    .ttl-input { width: 90px; }
    .ttl-unit  { font-size: 0.775rem; color: #555; }
    .cache-actions { display: flex; gap: 0.4rem; flex-wrap: wrap; margin-top: 0.75rem; }
    .toggle-row { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.6rem; }
    .toggle-label { width: 150px; font-size: 0.825rem; color: #bbb; flex-shrink: 0; }
    .toggle { position: relative; display: inline-block; width: 40px; height: 22px; flex-shrink: 0; }
    .toggle input { opacity: 0; width: 0; height: 0; }
    .toggle-slider {
      position: absolute; cursor: pointer; inset: 0;
      background: #2a2a2a; border-radius: 22px; transition: background 0.2s;
    }
    .toggle-slider::before {
      content: ""; position: absolute;
      width: 16px; height: 16px; left: 3px; bottom: 3px;
      background: #555; border-radius: 50%; transition: transform 0.2s, background 0.2s;
    }
    .toggle input:checked + .toggle-slider { background: #2d6a2d; }
    .toggle input:checked + .toggle-slider::before { transform: translateX(18px); background: #4caf50; }
    .notice { font-size: 0.75rem; color: #555; margin-top: 0.4rem; }

    /* ── Addon row expansion (per-addon settings) ── */
    .dnd-item-wrap { display: flex; flex-direction: column; gap: 0.3rem; }
    .dnd-item.expanded { border-color: #3a3a3a; }
    .dnd-settings {
      background: #131313; border: 1px solid #232323; border-top: none;
      border-radius: 0 0 8px 8px; padding: 0.6rem 0.75rem;
      margin-top: -0.3rem;
      display: flex; flex-wrap: wrap; gap: 0.75rem; align-items: center;
    }
    .dnd-settings label { font-size: 0.75rem; color: #999; display: flex; align-items: center; gap: 0.4rem; }
    .dnd-settings select, .dnd-settings input[type="number"] {
      background: #0f0f0f; border: 1px solid #2a2a2a; color: #e0e0e0;
      border-radius: 5px; padding: 0.25rem 0.4rem; font-size: 0.78rem; outline: none;
    }
    .btn-expand {
      background: none; border: 1px solid transparent; color: #555;
      border-radius: 6px; padding: 3px 8px; font-size: 0.78rem;
      cursor: pointer; white-space: nowrap; flex-shrink: 0;
    }
    .btn-expand:hover { border-color: #333; color: #aaa; }

    /* ── Rate limit profile rows ── */
    .rl-row {
      display: grid; grid-template-columns: 1fr 90px 90px 90px auto;
      gap: 0.5rem; align-items: center;
      background: #1a1a1a; border: 1px solid #242424; border-radius: 8px;
      padding: 0.4rem 0.6rem; margin-bottom: 0.35rem;
    }
    .rl-row input { width: 100%; }
    .rl-name-input { font-weight: 500; }
    .rl-head {
      display: grid; grid-template-columns: 1fr 90px 90px 90px 60px;
      gap: 0.5rem; font-size: 0.7rem; color: #555;
      padding: 0 0.6rem; margin-bottom: 0.25rem;
      text-transform: uppercase; letter-spacing: 0.04em;
    }
    .rl-add-form { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-top: 0.6rem; }
    .rl-add-form input { width: 110px; }
    .rl-add-form .rl-name-input { width: auto; flex: 1; min-width: 140px; }

    /* Exclude word tags */
    .exclude-tags { display: flex; flex-wrap: wrap; gap: 0.3rem; min-height: 1.5rem; }
    .exclude-tag {
      display: inline-flex; align-items: center; gap: 0.3rem;
      background: #2a1a1a; border: 1px solid #5a2a2a; border-radius: 20px;
      padding: 0.2rem 0.55rem; font-size: 0.775rem; color: #c07070;
    }
    .exclude-tag button {
      background: none; border: none; color: #883322; cursor: pointer;
      font-size: 0.9rem; line-height: 1; padding: 0; transition: color 0.15s;
    }
    .exclude-tag button:hover { color: #c0392b; }
  </style>
</head>
<body>
<div class="container">

  <!-- Header -->
  <div class="page-header">
    <div class="page-header-text">
      <h1>Stream Resolver</h1>
      <div class="subtitle">Aggregate, rank, and cast streams from multiple Stremio addons</div>
    </div>
    <span class="manifest-pill" id="manifest-url"></span>
    <a class="btn-install" id="install-link" href="#">Install in Stremio</a>
  </div>

  <!-- Source Addons -->
  <div class="card">
    <h2>Source Addons</h2>
    <p class="card-desc">Drag to set priority — top row is tried first. Streams from all addons are merged and ranked before selection.</p>
    <div class="dnd-list" id="addon-list"><p class="empty">Loading…</p></div>
    <div class="add-form">
      <div class="form-row">
        <input id="in-url"     type="text"   class="url-input"     placeholder="Manifest URL" />
        <input id="in-name"    type="text"   class="name-input"    placeholder="Name (optional)" />
        <input id="in-timeout" type="number" class="timeout-input" placeholder="Timeout ms" value="8000" min="500" />
        <button class="btn-add" onclick="addAddon()">Add</button>
      </div>
      <p class="status" id="addon-status"></p>
    </div>
  </div>

  <!-- Meta Addons -->
  <div class="card">
    <h2>Meta Addons</h2>
    <p class="card-desc">Used to look up runtime for probing validation. Queried in order — first match wins. Must advertise the <code>meta</code> resource (e.g. Cinemeta).</p>
    <div class="dnd-list" id="meta-addon-list"><p class="empty">Loading…</p></div>
    <div class="add-form">
      <div class="form-row">
        <input id="meta-in-url"     type="text"   class="url-input"     placeholder="Manifest URL" />
        <input id="meta-in-name"    type="text"   class="name-input"    placeholder="Name (optional)" />
        <input id="meta-in-timeout" type="number" class="timeout-input" placeholder="Timeout ms" value="5000" min="500" />
        <button class="btn-add" onclick="addMetaAddon()">Add</button>
      </div>
      <p class="notice">Cinemeta: <code>https://v3-cinemeta.strem.io</code></p>
      <p class="status" id="meta-status"></p>
    </div>
  </div>

  <!-- Streaming Defaults -->
  <div class="card">
    <h2>Streaming Defaults</h2>
    <p class="card-desc">Preferred streams are tried first (pass 1). If all fail, fallback streams are tried (pass 2). Excluded streams are never used. Per-request query params override preferred/fallback but config exclusions always apply.</p>

    <div class="section-label">Quality</div>
    <div class="bucket-rows" id="quality-rows">
      <div class="bucket-row" data-dim="quality" data-key="4K">
        <span class="bucket-row-label">4K</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="quality" data-key="1080p">
        <span class="bucket-row-label">1080p</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="quality" data-key="720p">
        <span class="bucket-row-label">720p</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="quality" data-key="480p">
        <span class="bucket-row-label">480p</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
    </div>

    <div class="section-label">Source Type</div>
    <div class="bucket-rows" id="sourcetype-rows">
      <div class="bucket-row" data-dim="source_type" data-key="remux">
        <span class="bucket-row-label">Remux</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="bluray">
        <span class="bucket-row-label">BluRay</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="web-dl">
        <span class="bucket-row-label">WEB-DL</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="webrip">
        <span class="bucket-row-label">WEBRip</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="hdtv">
        <span class="bucket-row-label">HDTV</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="dvd">
        <span class="bucket-row-label">DVD</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="source_type" data-key="cam">
        <span class="bucket-row-label">CAM / TS</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
    </div>

    <div class="section-label">Addon Source</div>
    <div class="bucket-rows" id="source-rows"><p class="bucket-empty">Loading…</p></div>
    <p class="notice" style="margin-bottom:0.75rem;">Populated from your configured source addons.</p>

    <div class="section-label">Audio Language</div>
    <div class="bucket-rows" id="audiolang-rows">
      <div class="bucket-row" data-dim="audio_lang" data-key="en">
        <span class="bucket-row-label">English</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="multi">
        <span class="bucket-row-label">Multi</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="fr">
        <span class="bucket-row-label">French</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="de">
        <span class="bucket-row-label">German</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="es">
        <span class="bucket-row-label">Spanish</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="pt">
        <span class="bucket-row-label">Portuguese</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="it">
        <span class="bucket-row-label">Italian</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="ja">
        <span class="bucket-row-label">Japanese</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="ko">
        <span class="bucket-row-label">Korean</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="zh">
        <span class="bucket-row-label">Chinese</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="ru">
        <span class="bucket-row-label">Russian</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="ar">
        <span class="bucket-row-label">Arabic</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
      <div class="bucket-row" data-dim="audio_lang" data-key="hi">
        <span class="bucket-row-label">Hindi</span>
        <div class="seg-group">
          <button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>
          <button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>
          <button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>
        </div>
      </div>
    </div>

    <div class="section-label">Size &amp; Bitrate Filters</div>
    <div class="filter-grid">
      <div class="filter-field">
        <span class="field-label">Min Size (GB)</span>
        <input type="number" id="def-min-size" min="0" step="0.1" placeholder="0 = no limit" />
      </div>
      <div class="filter-field">
        <span class="field-label">Max Size (GB)</span>
        <input type="number" id="def-max-size" min="0" step="0.1" placeholder="0 = no limit" />
      </div>
      <div class="filter-field">
        <span class="field-label">Min Bitrate (Mbps)</span>
        <input type="number" id="def-min-bitrate" min="0" step="0.1" placeholder="0 = no limit" />
      </div>
      <div class="filter-field">
        <span class="field-label">Max Bitrate (Mbps)</span>
        <input type="number" id="def-max-bitrate" min="0" step="0.1" placeholder="0 = no limit" />
      </div>
    </div>

    <div class="section-label">Excluded Words</div>
    <p class="card-desc" style="margin-bottom:0.5rem;">Streams whose name, title, or description contains any of these words (case-insensitive) are always skipped.</p>
    <div class="exclude-tags" id="exclude-tags"></div>
    <div class="form-row" style="margin-top:0.4rem;">
      <input type="text" id="exclude-input" class="url-input" placeholder="Type a word and press Enter or comma…" onkeydown="excludeKeydown(event)" />
    </div>

    <div style="margin-top:1rem;">
      <button class="btn-save" onclick="saveDefaults()">Save Defaults</button>
    </div>
    <p class="status" id="defaults-status"></p>
  </div>

  <!-- Cache -->
  <div class="card">
    <h2>Cache</h2>
    <div class="ttl-row">
      <span class="ttl-label">Stream list TTL</span>
      <input id="ttl-streams" class="ttl-input" type="number" min="0" placeholder="300" />
      <span class="ttl-unit">seconds</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Play result TTL</span>
      <input id="ttl-play" class="ttl-input" type="number" min="0" placeholder="3600" />
      <span class="ttl-unit">seconds</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Meta (runtime) TTL</span>
      <input id="ttl-meta" class="ttl-input" type="number" min="0" placeholder="86400" />
      <span class="ttl-unit">seconds</span>
    </div>
    <div style="display:flex;gap:0.5rem;align-items:center;margin-top:0.75rem;flex-wrap:wrap;">
      <button class="btn-save" onclick="saveCacheTTLs()">Save TTLs</button>
    </div>
    <div class="cache-actions">
      <button class="btn-clear" onclick="clearCache('streams')">Clear stream cache</button>
      <button class="btn-clear" onclick="clearCache('play')">Clear play cache</button>
      <button class="btn-clear" onclick="clearCache('meta')">Clear meta cache</button>
      <button class="btn-clear" onclick="clearCache('all')">Clear all</button>
    </div>
    <p class="status" id="cache-status"></p>
  </div>

  <!-- Rate Limit Profiles -->
  <div class="card">
    <h2>Rate Limit Profiles</h2>
    <p class="card-desc">Shared buckets for probe traffic. Addons referencing the same profile share its limits — e.g. AIOStreams and Meteor both backed by Torbox should share one profile. <code>0</code> in any field disables that dimension.</p>
    <div class="rl-head">
      <span>Name</span>
      <span>per min</span>
      <span>per hour</span>
      <span>concurrent</span>
      <span></span>
    </div>
    <div id="rl-list"><p class="empty">No profiles configured.</p></div>
    <div style="display:flex;gap:0.4rem;margin:0.5rem 0;">
      <button class="btn-save" onclick="saveRateLimitProfiles()">Save changes</button>
    </div>
    <div class="rl-add-form">
      <input id="rl-new-name" type="text" class="rl-name-input" placeholder="Profile name (e.g. torbox)" />
      <input id="rl-new-min" type="number" min="0" placeholder="per min" />
      <input id="rl-new-hour" type="number" min="0" placeholder="per hour" />
      <input id="rl-new-conc" type="number" min="0" placeholder="concurrent" />
      <button class="btn-add" onclick="addRateLimitProfile()">Add</button>
    </div>
    <p class="status" id="rl-status"></p>
  </div>

  <!-- Probing -->
  <div class="card">
    <h2>Probing</h2>
    <p class="card-desc">ffprobe validates stream duration before redirect. Combined with the probe cache and rate-limit profiles, this stays friendly with debrid API quotas.</p>
    <div class="toggle-row">
      <span class="toggle-label">Enable probing</span>
      <label class="toggle">
        <input type="checkbox" id="probing-enabled" />
        <span class="toggle-slider"></span>
      </label>
      <span style="font-size:0.775rem;color:#555;">Uses ffprobe to validate stream duration before playback</span>
    </div>
    <div class="toggle-row">
      <span class="toggle-label">Early exit</span>
      <label class="toggle">
        <input type="checkbox" id="probing-early-exit" />
        <span class="toggle-slider"></span>
      </label>
      <span style="font-size:0.775rem;color:#555;">Return as soon as the best stream passes — don't wait for slower probes</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Max candidates</span>
      <input id="probing-max-candidates" class="ttl-input" type="number" min="1" placeholder="5" />
      <span class="ttl-unit">top-N to probe per pass</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Probe timeout</span>
      <input id="probing-timeout" class="ttl-input" type="number" min="1" placeholder="15000" />
      <span class="ttl-unit">ms</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Cache success TTL</span>
      <input id="probing-cache-success" class="ttl-input" type="number" min="0" placeholder="3600" />
      <span class="ttl-unit">seconds — how long passing probes are reused</span>
    </div>
    <div class="ttl-row">
      <span class="ttl-label">Cache failure TTL</span>
      <input id="probing-cache-failure" class="ttl-input" type="number" min="0" placeholder="600" />
      <span class="ttl-unit">seconds — how long failed probes stay cached</span>
    </div>
    <div style="margin-top:0.75rem;">
      <button class="btn-save" onclick="saveProbingConfig()">Save</button>
    </div>
    <p class="status" id="probing-status"></p>
  </div>

  <!-- Raw Config JSON -->
  <div class="card">
    <div style="display:flex;align-items:center;justify-content:space-between;">
      <h2 style="margin-bottom:0;">Raw Config JSON</h2>
      <button class="btn-clear" id="raw-config-toggle" onclick="toggleRawConfig()" style="font-size:0.775rem;">Show</button>
    </div>
    <div id="raw-config-body" style="display:none;margin-top:0.9rem;">
      <p class="card-desc">Direct read/write of the full <code>config.json</code>. Edits here override all structured sections above on save.</p>
      <textarea id="raw-config" spellcheck="false" style="
        width:100%; min-height:340px; resize:vertical;
        background:#0a0a0a; border:1px solid #2a2a2a; color:#c8d3e0;
        border-radius:8px; padding:0.75rem; font-family:monospace; font-size:0.8rem;
        line-height:1.5; outline:none; box-sizing:border-box;
      "></textarea>
      <div style="display:flex;gap:0.5rem;align-items:center;margin-top:0.6rem;flex-wrap:wrap;">
        <button class="btn-save" onclick="saveRawConfig()">Save Config</button>
        <button class="btn-clear" onclick="loadRawConfig()">Reload</button>
      </div>
      <p class="status" id="raw-config-status"></p>
    </div>
  </div>

</div><!-- /container -->

<script>
  const BASE     = window.location.origin;
  const API      = BASE + '/api/addons';
  const META_API = BASE + '/api/meta-addons';

  document.getElementById('manifest-url').textContent = BASE + '/manifest.json';
  document.getElementById('install-link').href =
    'stremio://' + BASE.replace(/^https?:\/\//, '') + '/manifest.json';

  // ── Generic drag-and-drop for ordered lists ────────────────────────────────

  function wireDnd(item, container, onReorder) {
    let _src = null;
    item.draggable = true;
    item.addEventListener('dragstart', e => {
      _src = item;
      e.dataTransfer.effectAllowed = 'move';
      setTimeout(() => item.classList.add('dragging'), 0);
    });
    item.addEventListener('dragend', async () => {
      _src = null;
      item.classList.remove('dragging');
      container.querySelectorAll('.drag-over').forEach(i => i.classList.remove('drag-over'));
      if (onReorder) await onReorder();
    });
    item.addEventListener('dragover', e => {
      e.preventDefault();
      if (item !== _getActiveDrag()) item.classList.add('drag-over');
    });
    item.addEventListener('dragleave', () => item.classList.remove('drag-over'));
    item.addEventListener('drop', e => {
      e.preventDefault();
      const src = _getActiveDrag();
      if (src && src !== item) {
        const all = [...container.querySelectorAll('.dnd-item')];
        const si = all.indexOf(src), di = all.indexOf(item);
        container.insertBefore(src, si < di ? item.nextSibling : item);
      }
      item.classList.remove('drag-over');
    });
  }

  // Track the currently dragged item across closure boundaries
  let _activeDrag = null;
  function _getActiveDrag() { return _activeDrag; }

  // Patch wireDnd to use shared _activeDrag
  function wireDndShared(item, container, onReorder) {
    item.draggable = true;
    item.addEventListener('dragstart', e => {
      _activeDrag = item;
      e.dataTransfer.effectAllowed = 'move';
      setTimeout(() => item.classList.add('dragging'), 0);
    });
    item.addEventListener('dragend', async () => {
      _activeDrag = null;
      item.classList.remove('dragging');
      container.querySelectorAll('.drag-over').forEach(i => i.classList.remove('drag-over'));
      if (onReorder) await onReorder();
    });
    item.addEventListener('dragover', e => {
      e.preventDefault();
      if (_activeDrag && item !== _activeDrag) item.classList.add('drag-over');
    });
    item.addEventListener('dragleave', () => item.classList.remove('drag-over'));
    item.addEventListener('drop', e => {
      e.preventDefault();
      if (_activeDrag && _activeDrag !== item) {
        const all = [...container.querySelectorAll('.dnd-item')];
        const si = all.indexOf(_activeDrag), di = all.indexOf(item);
        container.insertBefore(_activeDrag, si < di ? item.nextSibling : item);
      }
      item.classList.remove('drag-over');
    });
  }

  // ── Source Addons ──────────────────────────────────────────────────────────

  function _makeAddonRow(a) {
    const wrap = document.createElement('div');
    wrap.className = 'dnd-item-wrap';
    wrap.dataset.url = a.url;

    const row = document.createElement('div');
    row.className = 'dnd-item';
    row.dataset.url = a.url;
    const badges =
      (a.skip_probe ? '<span class="badge" style="background:#2a1a1a;color:#c07070;">skip probe</span>' : '') +
      (a.rate_limit_profile ? '<span class="badge" style="background:#1a2a3a;color:#7090c0;">' + esc(a.rate_limit_profile) + '</span>' : '');
    row.innerHTML =
      '<span class="dnd-handle">⠿</span>' +
      '<div class="dnd-info">' +
        '<div class="dnd-name">' + esc(a.name) + '</div>' +
        '<div class="dnd-url">'  + esc(a.url)  + '</div>' +
      '</div>' +
      badges +
      '<span class="badge">' + (a.timeout_ms || 8000) + 'ms</span>' +
      '<button class="btn-expand" onclick="toggleAddonSettings(\'' + esc(a.url) + '\')">Settings</button>' +
      '<button class="btn-remove" onclick="removeAddon(\'' + esc(a.url) + '\')">Remove</button>';

    const settings = document.createElement('div');
    settings.className = 'dnd-settings';
    settings.style.display = 'none';
    settings.dataset.url = a.url;
    settings.innerHTML =
      '<label>Rate-limit profile' +
        '<select data-field="rate_limit_profile">' +
          '<option value="">(none)</option>' +
        '</select>' +
      '</label>' +
      '<label>Skip probe ' +
        '<input type="checkbox" data-field="skip_probe" ' + (a.skip_probe ? 'checked' : '') + ' />' +
      '</label>' +
      '<label>Timeout (ms) ' +
        '<input type="number" min="500" data-field="timeout_ms" value="' + (a.timeout_ms || 8000) + '" />' +
      '</label>' +
      '<button class="btn-save" onclick="saveAddonSettings(\'' + esc(a.url) + '\')">Save</button>';

    wrap.appendChild(row);
    wrap.appendChild(settings);

    // Populate profile dropdown with current profiles + select current value.
    const sel = settings.querySelector('select[data-field="rate_limit_profile"]');
    Object.keys(_currentProfiles).sort().forEach(name => {
      const opt = document.createElement('option');
      opt.value = name;
      opt.textContent = name;
      if (a.rate_limit_profile === name) opt.selected = true;
      sel.appendChild(opt);
    });

    return wrap;
  }

  function toggleAddonSettings(url) {
    const wrap = document.querySelector('.dnd-item-wrap[data-url="' + cssEsc(url) + '"]');
    if (!wrap) return;
    const settings = wrap.querySelector('.dnd-settings');
    const row = wrap.querySelector('.dnd-item');
    const isOpen = settings.style.display !== 'none';
    settings.style.display = isOpen ? 'none' : '';
    row.classList.toggle('expanded', !isOpen);
  }

  function cssEsc(s) { return String(s).replace(/(["\\])/g, '\\$1'); }

  async function saveAddonSettings(url) {
    const wrap = document.querySelector('.dnd-item-wrap[data-url="' + cssEsc(url) + '"]');
    if (!wrap) return;
    const body = {
      rate_limit_profile: wrap.querySelector('select[data-field="rate_limit_profile"]').value,
      skip_probe:         wrap.querySelector('input[data-field="skip_probe"]').checked,
      timeout_ms:         parseInt(wrap.querySelector('input[data-field="timeout_ms"]').value) || 8000,
    };
    const res = await fetch(API + '?url=' + encodeURIComponent(url), {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (res.ok) {
      setStatus('addon-status', 'Addon settings saved.', true);
      loadAddons();
    } else {
      setStatus('addon-status', 'Error: ' + (await res.text()).trim(), false);
    }
  }

  async function _saveAddonOrder() {
    const urls = [...document.querySelectorAll('#addon-list .dnd-item-wrap')].map(r => r.dataset.url);
    const res = await fetch(API, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(urls),
    });
    setStatus('addon-status', res.ok ? 'Order saved.' : 'Failed to save order.', res.ok);
  }

  async function loadAddons() {
    // Load profiles first so the per-addon settings panel can populate its dropdown.
    await loadRateLimitProfiles();
    const res = await fetch(API);
    const addons = await res.json();
    _refreshSourceRows(addons);
    const el = document.getElementById('addon-list');
    if (!addons || addons.length === 0) {
      el.innerHTML = '<p class="empty">No source addons configured yet.</p>';
      return;
    }
    el.innerHTML = '';
    addons.forEach(a => {
      const wrap = _makeAddonRow(a);
      wireDndShared(wrap, el, _saveAddonOrder);
      el.appendChild(wrap);
    });
  }

  async function addAddon() {
    const url        = document.getElementById('in-url').value.trim();
    const name       = document.getElementById('in-name').value.trim();
    const timeout_ms = parseInt(document.getElementById('in-timeout').value) || 8000;
    if (!url) { setStatus('addon-status', 'URL is required.', false); return; }
    const res = await fetch(API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url, name, timeout_ms }),
    });
    if (res.ok) {
      setStatus('addon-status', 'Addon added.', true);
      document.getElementById('in-url').value = '';
      document.getElementById('in-name').value = '';
      loadAddons();
    } else {
      setStatus('addon-status', 'Error: ' + (await res.text()).trim(), false);
    }
  }

  async function removeAddon(url) {
    if (!confirm('Remove ' + url + '?')) return;
    const res = await fetch(API + '?url=' + encodeURIComponent(url), { method: 'DELETE' });
    if (res.ok) loadAddons();
    else setStatus('addon-status', 'Failed to remove addon.', false);
  }

  // ── Meta Addons ────────────────────────────────────────────────────────────

  function _makeMetaRow(a, idx) {
    const row = document.createElement('div');
    row.className = 'dnd-item';
    row.dataset.url = a.url;
    row.innerHTML =
      '<span class="dnd-handle">⠿</span>' +
      '<div class="dnd-info">' +
        '<div class="dnd-name">' + esc(a.name) + '</div>' +
        '<div class="dnd-url">'  + esc(a.url)  + '</div>' +
      '</div>' +
      '<span class="badge">order ' + (idx + 1) + '</span>' +
      '<span class="badge">' + (a.timeout_ms || 5000) + 'ms</span>' +
      '<button class="btn-remove" onclick="removeMetaAddon(\'' + esc(a.url) + '\')">Remove</button>';
    return row;
  }

  async function _saveMetaOrder() {
    const urls = [...document.querySelectorAll('#meta-addon-list .dnd-item')].map(r => r.dataset.url);
    const res = await fetch(META_API, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(urls),
    });
    setStatus('meta-status', res.ok ? 'Order saved.' : 'Failed to save order.', res.ok);
  }

  async function loadMetaAddons() {
    const res = await fetch(META_API);
    const addons = await res.json();
    const el = document.getElementById('meta-addon-list');
    if (!addons || addons.length === 0) {
      el.innerHTML = '<p class="empty">No meta addons configured. Add Cinemeta to enable runtime validation.</p>';
      return;
    }
    el.innerHTML = '';
    addons.forEach((a, i) => {
      const row = _makeMetaRow(a, i);
      wireDndShared(row, el, _saveMetaOrder);
      el.appendChild(row);
    });
  }

  async function addMetaAddon() {
    const url        = document.getElementById('meta-in-url').value.trim();
    const name       = document.getElementById('meta-in-name').value.trim();
    const timeout_ms = parseInt(document.getElementById('meta-in-timeout').value) || 5000;
    if (!url) { setStatus('meta-status', 'URL is required.', false); return; }
    const res = await fetch(META_API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url, name, timeout_ms }),
    });
    if (res.ok) {
      setStatus('meta-status', 'Meta addon added.', true);
      document.getElementById('meta-in-url').value = '';
      document.getElementById('meta-in-name').value = '';
      loadMetaAddons();
    } else {
      setStatus('meta-status', 'Error: ' + (await res.text()).trim(), false);
    }
  }

  async function removeMetaAddon(url) {
    if (!confirm('Remove ' + url + '?')) return;
    const res = await fetch(META_API + '?url=' + encodeURIComponent(url), { method: 'DELETE' });
    if (res.ok) loadMetaAddons();
    else setStatus('meta-status', 'Failed to remove meta addon.', false);
  }

  // ── Cache ──────────────────────────────────────────────────────────────────

  async function loadCacheTTLs() {
    const res = await fetch(BASE + '/api/config/cache');
    if (!res.ok) return;
    const cc = await res.json();
    document.getElementById('ttl-streams').value = cc.stream_list_ttl_seconds || '';
    document.getElementById('ttl-play').value    = cc.play_result_ttl_seconds || '';
    document.getElementById('ttl-meta').value    = cc.meta_ttl_seconds        || '';
  }

  async function saveCacheTTLs() {
    const body = {
      stream_list_ttl_seconds: parseInt(document.getElementById('ttl-streams').value) || 0,
      play_result_ttl_seconds: parseInt(document.getElementById('ttl-play').value)    || 0,
      meta_ttl_seconds:        parseInt(document.getElementById('ttl-meta').value)    || 0,
    };
    const res = await fetch(BASE + '/api/config/cache', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    setStatus('cache-status', res.ok ? 'TTLs saved.' : 'Error: ' + (await res.text()).trim(), res.ok);
    if (res.ok) loadCacheTTLs();
  }

  async function clearCache(type) {
    const res = await fetch(BASE + '/api/cache?type=' + type, { method: 'DELETE' });
    setStatus('cache-status', res.ok ? 'Cleared ' + type + ' cache.' : 'Failed to clear cache.', res.ok);
  }

  // ── Probing ────────────────────────────────────────────────────────────────

  async function loadProbingConfig() {
    const res = await fetch(BASE + '/api/config/probing');
    if (!res.ok) return;
    const pc = await res.json();
    document.getElementById('probing-enabled').checked       = pc.enabled        || false;
    document.getElementById('probing-early-exit').checked    = pc.early_exit    !== false;
    document.getElementById('probing-max-candidates').value  = pc.max_candidates || '';
    document.getElementById('probing-timeout').value         = pc.timeout_ms     || '';
    document.getElementById('probing-cache-success').value   = pc.success_ttl_seconds || '';
    document.getElementById('probing-cache-failure').value   = pc.failure_ttl_seconds || '';
  }

  async function saveProbingConfig() {
    const body = {
      enabled:             document.getElementById('probing-enabled').checked,
      early_exit:          document.getElementById('probing-early-exit').checked,
      max_candidates:      parseInt(document.getElementById('probing-max-candidates').value) || 0,
      timeout_ms:          parseInt(document.getElementById('probing-timeout').value)        || 0,
      success_ttl_seconds: parseInt(document.getElementById('probing-cache-success').value)  || 0,
      failure_ttl_seconds: parseInt(document.getElementById('probing-cache-failure').value)  || 0,
    };
    // Drop zero values so the server doesn't reject them as "must be positive".
    Object.keys(body).forEach(k => {
      if (typeof body[k] === 'number' && body[k] === 0) delete body[k];
    });
    const res = await fetch(BASE + '/api/config/probing', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    setStatus('probing-status', res.ok ? 'Saved.' : 'Error: ' + (await res.text()).trim(), res.ok);
    if (res.ok) loadProbingConfig();
  }

  // ── Rate Limit Profiles ────────────────────────────────────────────────────

  let _currentProfiles = {};

  async function loadRateLimitProfiles() {
    const res = await fetch(BASE + '/api/config/rate-limits');
    if (!res.ok) {
      _currentProfiles = {};
      return;
    }
    _currentProfiles = await res.json() || {};
    _renderRateLimitList();
  }

  function _renderRateLimitList() {
    const el = document.getElementById('rl-list');
    const names = Object.keys(_currentProfiles).sort();
    if (names.length === 0) {
      el.innerHTML = '<p class="empty">No profiles configured.</p>';
      return;
    }
    el.innerHTML = '';
    names.forEach(name => {
      const p = _currentProfiles[name];
      const row = document.createElement('div');
      row.className = 'rl-row';
      row.dataset.name = name;
      row.innerHTML =
        '<input class="rl-name-input" type="text" value="' + esc(name) + '" data-field="name" />' +
        '<input type="number" min="0" value="' + (p.per_minute || 0) + '" data-field="per_minute" />' +
        '<input type="number" min="0" value="' + (p.per_hour || 0) + '" data-field="per_hour" />' +
        '<input type="number" min="0" value="' + (p.max_concurrent || 0) + '" data-field="max_concurrent" />' +
        '<button class="btn-remove" onclick="removeRateLimitProfile(\'' + esc(name) + '\')">×</button>';
      el.appendChild(row);
    });
  }

  function _collectRateLimitProfiles() {
    const next = {};
    document.querySelectorAll('#rl-list .rl-row').forEach(row => {
      const name = row.querySelector('input[data-field="name"]').value.trim();
      if (!name) return;
      next[name] = {
        per_minute:     parseInt(row.querySelector('input[data-field="per_minute"]').value)     || 0,
        per_hour:       parseInt(row.querySelector('input[data-field="per_hour"]').value)       || 0,
        max_concurrent: parseInt(row.querySelector('input[data-field="max_concurrent"]').value) || 0,
      };
    });
    return next;
  }

  async function addRateLimitProfile() {
    const name = document.getElementById('rl-new-name').value.trim();
    if (!name) { setStatus('rl-status', 'Profile name is required.', false); return; }
    const next = _collectRateLimitProfiles();
    next[name] = {
      per_minute:     parseInt(document.getElementById('rl-new-min').value)  || 0,
      per_hour:       parseInt(document.getElementById('rl-new-hour').value) || 0,
      max_concurrent: parseInt(document.getElementById('rl-new-conc').value) || 0,
    };
    await _saveRateLimitProfiles(next);
    document.getElementById('rl-new-name').value = '';
    document.getElementById('rl-new-min').value  = '';
    document.getElementById('rl-new-hour').value = '';
    document.getElementById('rl-new-conc').value = '';
  }

  async function removeRateLimitProfile(name) {
    if (!confirm('Remove profile "' + name + '"? Addons referencing it will lose their limit.')) return;
    const next = _collectRateLimitProfiles();
    delete next[name];
    await _saveRateLimitProfiles(next);
  }

  // Save all rate limit rows including in-place renames / edits.
  async function saveRateLimitProfiles() {
    await _saveRateLimitProfiles(_collectRateLimitProfiles());
  }

  async function _saveRateLimitProfiles(next) {
    const res = await fetch(BASE + '/api/config/rate-limits', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(next),
    });
    if (res.ok) {
      _currentProfiles = await res.json() || {};
      _renderRateLimitList();
      setStatus('rl-status', 'Profiles saved.', true);
      loadAddons(); // refresh dropdowns
    } else {
      setStatus('rl-status', 'Error: ' + (await res.text()).trim(), false);
    }
  }

  // ── Bucket rows ───────────────────────────────────────────────────────────

  function segClick(btn) {
    const group = btn.closest('.seg-group');
    group.querySelectorAll('.seg-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
  }

  function getBucket(dim) {
    const preferred = [], excluded = [];
    document.querySelectorAll('.bucket-row[data-dim="' + dim + '"]').forEach(row => {
      const key = row.dataset.key;
      const active = row.querySelector('.seg-btn.active');
      const val = active ? active.dataset.val : 'fallback';
      if (val === 'preferred') preferred.push(key);
      else if (val === 'excluded') excluded.push(key);
    });
    return { preferred, excluded };
  }

  function setBucket(dim, data) {
    document.querySelectorAll('.bucket-row[data-dim="' + dim + '"]').forEach(row => {
      const key = row.dataset.key;
      let val = 'fallback';
      if ((data.preferred || []).some(k => k.toLowerCase() === key.toLowerCase())) val = 'preferred';
      else if ((data.excluded || []).some(k => k.toLowerCase() === key.toLowerCase())) val = 'excluded';
      row.querySelectorAll('.seg-btn').forEach(b => {
        b.classList.toggle('active', b.dataset.val === val);
      });
    });
  }

  function _makeBucketRow(dim, key, label) {
    const row = document.createElement('div');
    row.className = 'bucket-row';
    row.dataset.dim = dim;
    row.dataset.key = key;
    row.innerHTML =
      '<span class="bucket-row-label">' + esc(label) + '</span>' +
      '<div class="seg-group">' +
        '<button class="seg-btn" data-val="preferred" onclick="segClick(this)">Prefer</button>' +
        '<button class="seg-btn active" data-val="fallback" onclick="segClick(this)">—</button>' +
        '<button class="seg-btn" data-val="excluded" onclick="segClick(this)">Exclude</button>' +
      '</div>';
    return row;
  }

  function _refreshSourceRows(addons) {
    const el = document.getElementById('source-rows');
    // Remember current state before rebuilding
    const current = getBucket('source');
    el.innerHTML = '';
    if (!addons || addons.length === 0) {
      el.innerHTML = '<p class="bucket-empty">No addons configured yet.</p>';
      return;
    }
    addons.forEach(a => el.appendChild(_makeBucketRow('source', a.name, a.name)));
    setBucket('source', current);
  }

  // ── Exclude words tag input ────────────────────────────────────────────────

  function _addExcludeTag(word) {
    word = word.trim();
    if (!word) return;
    const existing = [...document.querySelectorAll('#exclude-tags .exclude-tag')]
      .map(t => t.dataset.word);
    if (existing.includes(word.toLowerCase())) return;
    const tag = document.createElement('span');
    tag.className = 'exclude-tag';
    tag.dataset.word = word.toLowerCase();
    tag.innerHTML = esc(word) + '<button onclick="this.parentElement.remove()" title="Remove">×</button>';
    document.getElementById('exclude-tags').appendChild(tag);
  }

  function excludeKeydown(e) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      const val = e.target.value.replace(/,$/, '').trim();
      _addExcludeTag(val);
      e.target.value = '';
    }
  }

  function _getExcludeWords() {
    return [...document.querySelectorAll('#exclude-tags .exclude-tag')].map(t => t.dataset.word);
  }

  function _setExcludeWords(words) {
    document.getElementById('exclude-tags').innerHTML = '';
    (words || []).forEach(_addExcludeTag);
  }

  // ── Streaming Defaults — load / save ───────────────────────────────────────

  async function loadDefaults() {
    const res = await fetch(BASE + '/api/config/defaults');
    if (!res.ok) return;
    const d = await res.json();
    setBucket('quality',     d.quality     || {});
    setBucket('source_type', d.source_type || {});
    setBucket('source',      d.source      || {});
    setBucket('audio_lang',  d.audio_lang  || {});
    document.getElementById('def-min-size').value    = d.min_size    || '';
    document.getElementById('def-max-size').value    = d.max_size    || '';
    document.getElementById('def-min-bitrate').value = d.min_bitrate || '';
    document.getElementById('def-max-bitrate').value = d.max_bitrate || '';
    _setExcludeWords(Array.isArray(d.exclude_words) ? d.exclude_words : []);
  }

  async function saveDefaults() {
    const body = {
      quality:      getBucket('quality'),
      source_type:  getBucket('source_type'),
      source:       getBucket('source'),
      audio_lang:   getBucket('audio_lang'),
      min_size:     parseFloat(document.getElementById('def-min-size').value)    || 0,
      max_size:     parseFloat(document.getElementById('def-max-size').value)    || 0,
      min_bitrate:  parseFloat(document.getElementById('def-min-bitrate').value) || 0,
      max_bitrate:  parseFloat(document.getElementById('def-max-bitrate').value) || 0,
      exclude_words: _getExcludeWords(),
    };
    const res = await fetch(BASE + '/api/config/defaults', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    setStatus('defaults-status', res.ok ? 'Saved.' : 'Error: ' + (await res.text()).trim(), res.ok);
    if (res.ok) loadDefaults();
  }

  // ── Shared ─────────────────────────────────────────────────────────────────

  function setStatus(id, msg, ok) {
    const el = document.getElementById(id);
    el.textContent = msg;
    el.className = 'status ' + (ok ? 'ok' : 'err');
  }

  function esc(s) {
    return String(s)
      .replace(/&/g,'&amp;').replace(/</g,'&lt;')
      .replace(/>/g,'&gt;').replace(/"/g,'&quot;')
      .replace(/'/g,'&#39;');
  }

  // ── Raw Config JSON ────────────────────────────────────────────────────────

  let _rawConfigLoaded = false;

  function toggleRawConfig() {
    const body = document.getElementById('raw-config-body');
    const btn  = document.getElementById('raw-config-toggle');
    const open = body.style.display === 'none';
    body.style.display = open ? '' : 'none';
    btn.textContent = open ? 'Hide' : 'Show';
    if (open && !_rawConfigLoaded) { _rawConfigLoaded = true; loadRawConfig(); }
  }

  async function loadRawConfig() {
    const res = await fetch(BASE + '/api/config');
    if (!res.ok) { setStatus('raw-config-status', 'Failed to load config.', false); return; }
    const text = await res.text();
    document.getElementById('raw-config').value = text.trimEnd();
    setStatus('raw-config-status', '', true);
  }

  async function saveRawConfig() {
    const text = document.getElementById('raw-config').value.trim();
    let parsed;
    try { parsed = JSON.parse(text); } catch (e) {
      setStatus('raw-config-status', 'Invalid JSON: ' + e.message, false);
      return;
    }
    const res = await fetch(BASE + '/api/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(parsed, null, 2),
    });
    if (res.ok) {
      const saved = await res.text();
      document.getElementById('raw-config').value = saved.trimEnd();
      setStatus('raw-config-status', 'Config saved.', true);
      loadRateLimitProfiles(); loadAddons(); loadMetaAddons(); loadCacheTTLs(); loadProbingConfig(); loadDefaults();
    } else {
      setStatus('raw-config-status', 'Error: ' + (await res.text()).trim(), false);
    }
  }

  loadAddons();
  loadMetaAddons();
  loadCacheTTLs();
  loadProbingConfig();
  loadDefaults();
  // loadAddons already loads rate limit profiles first; no separate init needed.
</script>
</body>
</html>`
}
