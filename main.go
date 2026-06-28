// Command portfolio is the single-binary backend for the portfolio site.
//
// It runs two independent HTTP listeners:
//
//	:8080  PUBLIC surface — the embedded static site plus /api/*. This is the
//	       only port the cluster Ingress routes to.
//	:9090  PRIVATE surface — /metrics (scraped in-cluster by VictoriaMetrics)
//	       and /healthz (Kubernetes probes). Never exposed to the internet.
//
// Splitting the ports is what lets a single process serve everything while
// still keeping /metrics and /healthz off the public internet (the Ingress
// only forwards :8080).
//
// At runtime it reads live data from VictoriaMetrics and Gitea over the
// cluster-internal network, holds no credentials, and exposes its own metrics
// so it shows up as a monitored service in the very stack it reports on.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.henrydowd.dev/henry/portfolio/internal/api"
	"git.henrydowd.dev/henry/portfolio/internal/config"
	"git.henrydowd.dev/henry/portfolio/internal/gitea"
	"git.henrydowd.dev/henry/portfolio/internal/metrics"
	"git.henrydowd.dev/henry/portfolio/internal/vm"
)

// version and commit are injected at build time via:
//
//	-ldflags "-X main.version=<tag> -X main.commit=<sha>"
//
// (see the Dockerfile). They default to dev values for local `go run`.
var (
	version = "dev"
	commit  = "none"
)

// commitLimit is how many recent commits the activity panel shows.
const commitLimit = 6

func main() {
	cfg := config.Load()
	log.Printf("portfolio %s (%s) starting: public=%s metrics=%s vm=%s gitea=%s repo=%s",
		version, commit, cfg.Addr, cfg.MetricsAddr, cfg.VMURL, cfg.GiteaURL, cfg.GiteaRepo)

	m := metrics.New(version, commit)

	// Data clients (cluster-internal, no auth) and the API layer that caches
	// and curates their output into JSON for the front-end.
	vmClient := vm.New(cfg.VMURL, cfg.RequestTimeout)
	giteaClient := gitea.New(cfg.GiteaURL, cfg.RequestTimeout)
	apiSrv := api.New(version, vmClient, giteaClient, cfg.GiteaRepo, commitLimit, cfg.CacheTTL, cfg.RequestTimeout, m)

	// Public mux: /api first, then the embedded site as the catch-all. Wrapped
	// in request instrumentation so every public request is counted/timed.
	public := http.NewServeMux()
	apiSrv.Routes(public)
	public.Handle("/", staticHandler())

	publicServer := &http.Server{
		Addr: cfg.Addr,
		// Security headers wrap the instrumented public mux. The private metrics
		// listener below is deliberately left without them.
		Handler:           securityHeaders(m.Middleware(public)),
		ReadHeaderTimeout: 5 * time.Second, // basic slow-loris protection
	}

	// Private mux: metrics + health. Not instrumented (we don't want scrapes in
	// our own request metrics) and never Ingress-routed.
	private := http.NewServeMux()
	private.Handle("/metrics", m.Handler())
	private.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	metricsServer := &http.Server{
		Addr:              cfg.MetricsAddr,
		Handler:           private,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Keep the upstream snapshots warm independently of public traffic: probe
	// VictoriaMetrics and Gitea on the cache cadence, starting immediately. This
	// makes portfolio_upstream_up present and current from process start — the
	// Kubernetes probes hit /healthz on :9090, not /api, so an idle pod would
	// otherwise never populate it — and means the first real visitor already gets
	// live data instead of triggering a cold load.
	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()
	go warmUpstreams(runCtx, apiSrv, cfg.CacheTTL)

	// Run both servers. The first to return an error trips shutdown.
	errCh := make(chan error, 2)
	go func() { errCh <- serve(publicServer) }()
	go func() { errCh <- serve(metricsServer) }()

	// Block until a server fails fatally or the platform asks us to stop
	// (Kubernetes sends SIGTERM before SIGKILL on pod termination).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
	case sig := <-sigCh:
		log.Printf("received %s, shutting down", sig)
	}

	// Graceful shutdown: stop the warmer, then stop accepting new connections and
	// let in-flight requests drain (up to the deadline) so a rollout doesn't cut
	// requests.
	cancelRun()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = publicServer.Shutdown(ctx)
	_ = metricsServer.Shutdown(ctx)
	log.Print("stopped")
}

// serve runs s and treats the clean-shutdown sentinel as success.
func serve(s *http.Server) error {
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// warmUpstreams probes the cached upstreams on the cache TTL, starting at once,
// until ctx is cancelled. Each probe reports upstream reachability to the
// metrics registry as a side effect (see api.API.Warm), so portfolio_upstream_up
// reflects real health on an idle pod instead of only after live /api traffic.
func warmUpstreams(ctx context.Context, a *api.API, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	a.Warm() // probe once at startup so the gauges are populated right away
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			a.Warm()
		}
	}
}
