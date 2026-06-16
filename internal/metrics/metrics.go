// Package metrics implements a tiny, dependency-free Prometheus exposition for
// the service's own telemetry.
//
// Why hand-rolled instead of prometheus/client_golang? The whole service is a
// single static binary with no third-party dependencies (see README.md); a
// handful of metrics in the Prometheus text exposition format keeps it that
// way. The format is simple and stable, so for this handful of series a
// third-party dependency isn't worth it.
//
// Exposed series (all prefixed portfolio_ except the conventional go_ one):
//
//	portfolio_build_info{version,commit}                gauge, always 1
//	portfolio_http_requests_total{route,method,code}    counter
//	portfolio_http_request_duration_seconds{route}      histogram
//	portfolio_http_requests_in_flight                   gauge
//	portfolio_upstream_up{upstream}                      gauge (1 reachable / 0 not)
//	go_goroutines                                        gauge
package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// durationBuckets are the histogram's upper bounds in seconds, tuned for a fast
// static+JSON service. A +Inf bucket (== total count) is emitted implicitly.
var durationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

// Metrics is the in-process registry. All access goes through mu.
type Metrics struct {
	version, commit string

	mu        sync.Mutex
	requests  map[reqKey]uint64     // request counter by label tuple
	durations map[string]*histogram // latency histogram by route
	inflight  int64                 // in-flight gauge
	upstream  map[string]float64    // upstream reachability by name (0/1)
}

type reqKey struct{ route, method, code string }

// New returns a registry stamped with the given build identifiers.
func New(version, commit string) *Metrics {
	return &Metrics{
		version:   version,
		commit:    commit,
		requests:  map[reqKey]uint64{},
		durations: map[string]*histogram{},
		upstream:  map[string]float64{},
	}
}

// SetUpstreamUp records whether an upstream (e.g. "vm", "gitea") was reachable
// on its most recent use. Called by the data loaders.
func (m *Metrics) SetUpstreamUp(name string, up bool) {
	v := 0.0
	if up {
		v = 1
	}
	m.mu.Lock()
	m.upstream[name] = v
	m.mu.Unlock()
}

// Middleware instruments the wrapped handler: in-flight gauge, request counter
// (by route/method/status) and latency histogram (by route). Paths are folded
// into a small set of route labels by classify to keep cardinality bounded.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.addInflight(1)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		m.addInflight(-1)
		m.observe(classify(r.URL.Path), r.Method, sw.status, time.Since(start).Seconds())
	})
}

// Handler renders the current metrics in the Prometheus text exposition format.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		m.mu.Lock()
		defer m.mu.Unlock()
		var b strings.Builder

		writeHelp(&b, "portfolio_build_info", "Build version and commit; constant 1.", "gauge")
		fmt.Fprintf(&b, "portfolio_build_info{version=%q,commit=%q} 1\n", m.version, m.commit)

		writeHelp(&b, "portfolio_http_requests_total", "Total HTTP requests handled.", "counter")
		for _, k := range sortedReqKeys(m.requests) {
			fmt.Fprintf(&b, "portfolio_http_requests_total{route=%q,method=%q,code=%q} %d\n",
				k.route, k.method, k.code, m.requests[k])
		}

		writeHelp(&b, "portfolio_http_request_duration_seconds", "HTTP request latency in seconds.", "histogram")
		for _, route := range sortedKeys(m.durations) {
			h := m.durations[route]
			var cumulative uint64
			for i, bound := range durationBuckets {
				cumulative += h.counts[i]
				fmt.Fprintf(&b, "portfolio_http_request_duration_seconds_bucket{route=%q,le=%q} %d\n",
					route, formatFloat(bound), cumulative)
			}
			fmt.Fprintf(&b, "portfolio_http_request_duration_seconds_bucket{route=%q,le=\"+Inf\"} %d\n", route, h.count)
			fmt.Fprintf(&b, "portfolio_http_request_duration_seconds_sum{route=%q} %s\n", route, formatFloat(h.sum))
			fmt.Fprintf(&b, "portfolio_http_request_duration_seconds_count{route=%q} %d\n", route, h.count)
		}

		writeHelp(&b, "portfolio_http_requests_in_flight", "Requests currently being served.", "gauge")
		fmt.Fprintf(&b, "portfolio_http_requests_in_flight %d\n", m.inflight)

		writeHelp(&b, "portfolio_upstream_up", "Whether an upstream was reachable on last use (1/0).", "gauge")
		for _, name := range sortedKeys(m.upstream) {
			fmt.Fprintf(&b, "portfolio_upstream_up{upstream=%q} %s\n", name, formatFloat(m.upstream[name]))
		}

		writeHelp(&b, "go_goroutines", "Number of goroutines that currently exist.", "gauge")
		fmt.Fprintf(&b, "go_goroutines %d\n", runtime.NumGoroutine())

		_, _ = w.Write([]byte(b.String()))
	})
}

func (m *Metrics) addInflight(delta int64) {
	m.mu.Lock()
	m.inflight += delta
	m.mu.Unlock()
}

func (m *Metrics) observe(route, method string, code int, seconds float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[reqKey{route, method, strconv.Itoa(code)}]++
	h := m.durations[route]
	if h == nil {
		h = newHistogram()
		m.durations[route] = h
	}
	h.observe(seconds)
}

// histogram accumulates observations into per-bucket counts. Buckets are made
// cumulative only at render time.
type histogram struct {
	counts []uint64 // counts[i] = observations whose value falls in bucket i
	sum    float64
	count  uint64
}

func newHistogram() *histogram {
	return &histogram{counts: make([]uint64, len(durationBuckets))}
}

func (h *histogram) observe(v float64) {
	h.sum += v
	h.count++
	for i, bound := range durationBuckets {
		if v <= bound {
			h.counts[i]++
			return
		}
	}
	// Values above the last bound are counted only in count (the +Inf bucket).
}

// classify maps a request path to a low-cardinality route label. Without it,
// every distinct URL (and every fingerprinted asset) would become its own
// time series.
func classify(path string) string {
	switch {
	case path == "/api/status":
		return "/api/status"
	case path == "/api/git":
		return "/api/git"
	case path == "/" || path == "/index.html":
		return "/"
	case strings.HasPrefix(path, "/_astro/"):
		return "/_astro/*"
	default:
		return "other"
	}
}

// statusWriter records the response status code for instrumentation.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.wroteHeader = true // an implicit 200 has now been sent
	return w.ResponseWriter.Write(b)
}

func writeHelp(b *strings.Builder, name, help, typ string) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s %s\n", name, help, name, typ)
}

// formatFloat renders a float the way Prometheus expects (shortest exact form).
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func sortedReqKeys(m map[reqKey]uint64) []reqKey {
	keys := make([]reqKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].code < keys[j].code
	})
	return keys
}

// sortedKeys returns the keys of any string-keyed map in sorted order, so the
// exposition output is stable between scrapes.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
