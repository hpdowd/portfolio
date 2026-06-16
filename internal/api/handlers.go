// Package api implements the JSON endpoints the front-end consumes:
// /api/status and /api/git.
//
// Both read from short-lived caches (so page traffic doesn't translate directly
// into upstream load) and both degrade gracefully when an upstream is
// unreachable — the handlers always return a valid document, with a "live"
// flag the page uses to show freshness.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"git.henrydowd.dev/henry/portfolio/internal/cache"
	"git.henrydowd.dev/henry/portfolio/internal/gitea"
	"git.henrydowd.dev/henry/portfolio/internal/vm"
)

// upstreamReporter is the part of *metrics.Metrics this package actually needs.
// Taking an interface keeps the metrics implementation out of this import graph.
type upstreamReporter interface {
	SetUpstreamUp(name string, up bool)
}

// API holds the cached data sources behind the JSON endpoints.
type API struct {
	started time.Time
	version string

	status *cache.TTL[vm.Snapshot]
	git    *cache.TTL[[]gitea.Commit]
}

// New wires the data clients and their caches.
//
// The cache loaders own their own timeouts (reqTimeout) and report each
// upstream's reachability to m, so the cache itself stays oblivious to metrics.
func New(
	version string,
	vmClient *vm.Client,
	giteaClient *gitea.Client,
	repo string,
	commitLimit int,
	ttl, reqTimeout time.Duration,
	m upstreamReporter,
) *API {
	statusLoader := func() (vm.Snapshot, error) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()
		s, err := vmClient.Fetch(ctx)
		m.SetUpstreamUp("vm", err == nil)
		return s, err
	}
	gitLoader := func() ([]gitea.Commit, error) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()
		cs, err := giteaClient.RecentCommits(ctx, repo, commitLimit)
		m.SetUpstreamUp("gitea", err == nil)
		return cs, err
	}

	return &API{
		started: time.Now(),
		version: version,
		status:  cache.New(ttl, statusLoader),
		git:     cache.New(ttl, gitLoader),
	}
}

// Routes registers the API handlers on mux.
func (a *API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/git", a.handleGit)
}

type statusResponse struct {
	Service struct {
		UptimeSeconds int64  `json:"uptime_seconds"`
		Version       string `json:"version"`
	} `json:"service"`
	Cluster *vm.Snapshot `json:"cluster"` // null until VictoriaMetrics responds at least once
	Live    bool         `json:"live"`    // is the cluster data fresh from VictoriaMetrics?
}

// handleStatus returns the service's own uptime (always available, in-process)
// plus the cluster snapshot from VictoriaMetrics (when reachable).
func (a *API) handleStatus(w http.ResponseWriter, _ *http.Request) {
	var resp statusResponse
	resp.Service.UptimeSeconds = int64(time.Since(a.started).Seconds())
	resp.Service.Version = a.version
	if snap, ok := a.status.Get(); ok {
		resp.Cluster = &snap
		resp.Live = true
	}
	writeJSON(w, resp)
}

type gitResponse struct {
	Commits []gitea.Commit `json:"commits"`
	Live    bool           `json:"live"`
}

// handleGit returns the recent commit feed (empty but valid until the first
// successful fetch).
func (a *API) handleGit(w http.ResponseWriter, _ *http.Request) {
	resp := gitResponse{Commits: []gitea.Commit{}}
	if commits, ok := a.git.Get(); ok {
		resp.Commits = commits
		resp.Live = true
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// A short client-side cache; the server already rate-limits upstream calls
	// via its own cache, so this just smooths bursts of page polling.
	w.Header().Set("Cache-Control", "public, max-age=15")
	_ = json.NewEncoder(w).Encode(v)
}
