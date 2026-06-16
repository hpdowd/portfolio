// Package config loads runtime configuration from environment variables.
//
// Every setting has a default that lets the binary run locally with no setup.
// In the cluster the Deployment overrides VM_URL and GITEA_URL with the
// cluster-internal Service addresses (see the homelab repo's
// k8s/apps/portfolio/deployment.yaml).
package config

import (
	"os"
	"time"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	Addr        string // public listener: static site + /api (behind the Ingress)
	MetricsAddr string // private listener: /metrics + /healthz (never Ingress-routed)

	VMURL     string // VictoriaMetrics base URL (cluster-internal, no auth)
	GiteaURL  string // Gitea base URL
	GiteaRepo string // "owner/name" whose commits feed the activity panel (must be public)

	CacheTTL       time.Duration // how long curated upstream data is reused before a refresh
	RequestTimeout time.Duration // per-call timeout for each upstream (VM/Gitea) request
}

// Load reads the environment and applies defaults.
func Load() Config {
	return Config{
		Addr:           getenv("PORTFOLIO_ADDR", ":8080"),
		MetricsAddr:    getenv("PORTFOLIO_METRICS_ADDR", ":9090"),
		VMURL:          getenv("VM_URL", "http://localhost:8428"),
		GiteaURL:       getenv("GITEA_URL", "https://git.henrydowd.dev"),
		GiteaRepo:      getenv("GITEA_REPO", "henry/homelab"),
		CacheTTL:       getdur("PORTFOLIO_CACHE_TTL", 30*time.Second),
		RequestTimeout: getdur("PORTFOLIO_REQUEST_TIMEOUT", 5*time.Second),
	}
}

// getenv returns the value of key, or def if it is unset or empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getdur parses key as a Go duration (e.g. "30s"), falling back to def if it is
// unset or unparseable.
func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
