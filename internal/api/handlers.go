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

// The upstreams whose reachability this package reports via
// portfolio_upstream_up. Named once so registration and reporting can't drift.
const (
	upstreamVM    = "vm"
	upstreamGitea = "gitea"
)

// upstreamReporter is the part of *metrics.Metrics this package actually needs.
// Taking an interface keeps the metrics implementation out of this import graph.
type upstreamReporter interface {
	SetUpstreamUp(name string, up bool)
	RegisterUpstreams(names ...string)
}

// API holds the cached data sources behind the JSON endpoints.
type API struct {
	started time.Time
	version string

	status *cache.TTL[vm.Snapshot]
	git    *cache.TTL[[]gitea.Commit]
	uptime *cache.TTL[vm.Uptime]
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
		m.SetUpstreamUp(upstreamVM, err == nil)
		return s, err
	}
	gitLoader := func() ([]gitea.Commit, error) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()
		cs, err := giteaClient.RecentCommits(ctx, repo, commitLimit)
		m.SetUpstreamUp(upstreamGitea, err == nil)
		return cs, err
	}

	// The 30-day availability history for the status page. Its own VM reachability
	// is already covered by statusLoader's upstreamVM gauge, so it doesn't report
	// separately. Not warmed (see Warm): it's lazy-loaded on the first /api/uptime
	// request, since the status page is visited far less than the home page.
	uptimeLoader := func() (vm.Uptime, error) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()
		return vmClient.FetchUptime(ctx)
	}

	// Seed both series at 0 so /metrics reports them from the first scrape,
	// before any loader has run (see RegisterUpstreams). The loaders flip them to
	// the observed value on each refresh.
	m.RegisterUpstreams(upstreamVM, upstreamGitea)

	return &API{
		started: time.Now(),
		version: version,
		status:  cache.New(ttl, statusLoader),
		git:     cache.New(ttl, gitLoader),
		uptime:  cache.New(ttl, uptimeLoader),
	}
}

// Routes registers the API handlers on mux.
func (a *API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/git", a.handleGit)
	mux.HandleFunc("/api/uptime", a.handleUptime)
}

// Warm refreshes every cached upstream so each one's reachability is observed
// (and reported to metrics) without waiting for live /api traffic. Get() probes
// in the background when the cached value is missing or stale; the returned
// values are intentionally discarded. Call it at startup and on the cache
// cadence to keep portfolio_upstream_up current on an idle pod — and so the
// first visitor already has fresh data to read.
func (a *API) Warm() {
	a.status.Get()
	a.git.Get()
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

type uptimeResponse struct {
	Uptime *vm.Uptime `json:"uptime"` // null until VictoriaMetrics answers the range query
	Live   bool       `json:"live"`
}

// handleUptime returns the platform availability history for the status page.
// Lazily loaded and cached; the first request may come back not-live while the
// background load runs, then fills on the next poll.
func (a *API) handleUptime(w http.ResponseWriter, _ *http.Request) {
	var resp uptimeResponse
	if u, ok := a.uptime.Get(); ok {
		resp.Uptime = &u
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
