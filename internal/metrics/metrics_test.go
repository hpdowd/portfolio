package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestClassify checks the path-to-route folding that keeps metric cardinality
// bounded (every fingerprinted asset must collapse to one /_astro/* label).
func TestClassify(t *testing.T) {
	tests := map[string]string{
		"/api/status":            "/api/status",
		"/api/git":               "/api/git",
		"/":                      "/",
		"/index.html":            "/",
		"/_astro/app.9f3c.js":    "/_astro/*",
		"/_astro/style.ab12.css": "/_astro/*",
		"/cv":                    "other",
		"/favicon.svg":           "other",
	}
	for path, want := range tests {
		if got := classify(path); got != want {
			t.Errorf("classify(%q) = %q, want %q", path, got, want)
		}
	}
}

// TestExposition drives one request through the middleware, then renders the
// registry and asserts the Prometheus text exposition contains the series the
// dashboard and alerts depend on. Bucket placement is timing-dependent, so the
// histogram is checked only via its always-true +Inf/count totals here.
func TestExposition(t *testing.T) {
	m := New("v1.2.3", "abcdef")
	m.SetUpstreamUp("vm", true)
	m.SetUpstreamUp("gitea", false)

	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/status", nil))

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain...", ct)
	}

	want := []string{
		`portfolio_build_info{version="v1.2.3",commit="abcdef"} 1`,
		`# TYPE portfolio_http_requests_total counter`,
		`portfolio_http_requests_total{route="/api/status",method="GET",code="200"} 1`,
		`portfolio_http_request_duration_seconds_bucket{route="/api/status",le="+Inf"} 1`,
		`portfolio_http_request_duration_seconds_count{route="/api/status"} 1`,
		`portfolio_http_requests_in_flight 0`,
		`portfolio_upstream_up{upstream="gitea"} 0`,
		`portfolio_upstream_up{upstream="vm"} 1`,
		`go_goroutines `,
	}
	for _, line := range want {
		if !strings.Contains(body, line) {
			t.Errorf("exposition missing:\n  %s\n--- got ---\n%s", line, body)
		}
	}
}

// TestHistogram checks the per-bucket accounting: a value above the last bound
// is counted only in the total (the implicit +Inf bucket), never in a sized one.
func TestHistogram(t *testing.T) {
	h := newHistogram()
	h.observe(0.002) // <= 0.005  -> bucket index 1
	h.observe(0.2)   // <= 0.25   -> bucket index 6
	h.observe(5)     // above 2.5 -> total only

	if h.count != 3 {
		t.Fatalf("count = %d, want 3", h.count)
	}
	if h.counts[1] != 1 || h.counts[6] != 1 {
		t.Errorf("bucket counts = %v, want index 1 and 6 set", h.counts)
	}
	var bucketed uint64
	for _, c := range h.counts {
		bucketed += c
	}
	if bucketed != 2 {
		t.Errorf("bucketed = %d, want 2 (the 5.0 lives only in +Inf)", bucketed)
	}
}
