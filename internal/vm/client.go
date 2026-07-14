// Package vm is a minimal client for VictoriaMetrics' Prometheus-compatible
// HTTP query API.
//
// It only ever issues the FIXED set of instant queries defined in Fetch — never
// any caller- or user-supplied PromQL — so there is no query-injection surface.
// VictoriaMetrics runs without authentication on the cluster-internal network,
// so no credentials are involved.
package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client queries a VictoriaMetrics (or Prometheus) HTTP API at base.
type Client struct {
	base string
	http *http.Client
}

// New returns a Client that talks to the API at base, applying timeout to each
// request.
func New(base string, timeout time.Duration) *Client {
	return &Client{base: base, http: &http.Client{Timeout: timeout}}
}

// Snapshot is the curated set of cluster numbers shown on the status panel.
type Snapshot struct {
	TargetsUp    int     `json:"targets_up"`
	TargetsTotal int     `json:"targets_total"`
	AlertsFiring int     `json:"alerts_firing"`
	RequestRate  float64 `json:"request_rate"` // this service's own req/s, scraped back from VM

	// Services maps a curated, safe canonical name (never a raw job/pod/IP label)
	// to whether that component's scrape targets are all up. Only components that
	// actually have a target in VictoriaMetrics appear, so the front-end lights a
	// diagram node iff it's genuinely monitored. Best-effort: nil if the per-job
	// query failed, without failing the rest of the snapshot.
	Services map[string]bool `json:"services,omitempty"`
}

// serviceMatchers maps VictoriaMetrics `job` labels to the safe canonical names
// the front-end knows, by case-insensitive substring. A canonical is reported
// up only if every matching job is up. Nothing but these canonical names ever
// leaves the process, so raw job/pod/instance labels are never exposed publicly.
var serviceMatchers = []struct {
	canonical string
	subs      []string
}{
	{"portfolio", []string{"portfolio"}},
	{"longhorn", []string{"longhorn"}},
	{"victoriametrics", []string{"victoria", "vmsingle", "vmagent", "vmalert"}},
	{"grafana", []string{"grafana"}},
	{"alertmanager", []string{"alertmanager"}},
	{"traefik", []string{"traefik"}},
	{"argocd", []string{"argocd", "argo-cd"}},
	{"gitea", []string{"gitea"}},
	{"nextcloud", []string{"nextcloud"}},
	{"immich", []string{"immich"}},
	{"collabora", []string{"collabora"}},
	{"kiwix", []string{"kiwix"}},
	{"file-parser", []string{"file-parser", "fileparser"}},
}

// curateServices folds raw job→up pairs into the canonical name→up map exposed
// to the front-end. Pure (no I/O) so it is unit-tested directly.
func curateServices(jobs map[string]bool) map[string]bool {
	out := map[string]bool{}
	for _, m := range serviceMatchers {
		found, up := false, true
		for job, jobUp := range jobs {
			lj := strings.ToLower(job)
			for _, sub := range m.subs {
				if strings.Contains(lj, sub) {
					found = true
					up = up && jobUp
					break
				}
			}
		}
		if found {
			out[m.canonical] = up
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Fetch runs the fixed queries that populate a Snapshot. Any single failed
// query fails the whole fetch; the caller degrades gracefully.
func (c *Client) Fetch(ctx context.Context) (Snapshot, error) {
	var s Snapshot
	var err error

	// Healthy vs total scrape targets. `up` is emitted for every target, so
	// these always have data once anything is being scraped.
	if s.TargetsUp, err = c.scalarInt(ctx, "count(up == 1)"); err != nil {
		return s, err
	}
	if s.TargetsTotal, err = c.scalarInt(ctx, "count(up)"); err != nil {
		return s, err
	}
	// Actionable firing alerts only. The severity filter drops the always-on
	// Watchdog (the alert-pipeline dead-man's switch, severity "none"); the
	// alertname filter drops the steady-state request-overcommit warnings this
	// small 2-node cluster always carries. So the count means "worth a look",
	// not background noise — which is what the live status badge keys on.
	// An empty result (nothing firing) yields 0.
	if s.AlertsFiring, err = c.scalarInt(ctx,
		`count(ALERTS{alertstate="firing",severity=~"warning|critical",alertname!~"KubeCPUOvercommit|KubeMemoryOvercommit"})`); err != nil {
		return s, err
	}
	// The service's own request rate — visible proof the recursive scrape works.
	// Empty (and therefore 0) until VictoriaMetrics has scraped this pod at least twice.
	if s.RequestRate, err = c.scalar(ctx, `sum(rate(portfolio_http_requests_total[5m]))`); err != nil {
		return s, err
	}
	// Per-component health for the diagram nodes. Best-effort: a failure here
	// leaves Services nil rather than blanking the whole panel.
	if jobs, jerr := c.jobUp(ctx); jerr == nil {
		s.Services = curateServices(jobs)
	}
	return s, nil
}

// jobUp returns each scrape job's aggregate health (up only if all of that job's
// targets are up), keyed by the raw `job` label. `min by (job) (up)` yields one
// series per job with value 1 (all up) or 0 (any down).
func (c *Client) jobUp(ctx context.Context) (map[string]bool, error) {
	samples, err := c.vector(ctx, "min by (job) (up)")
	if err != nil {
		return nil, err
	}
	jobs := make(map[string]bool, len(samples))
	for _, s := range samples {
		if job := s.metric["job"]; job != "" {
			jobs[job] = s.value == 1
		}
	}
	return jobs, nil
}

// scalarInt rounds the result of scalar to the nearest integer.
func (c *Client) scalarInt(ctx context.Context, query string) (int, error) {
	f, err := c.scalar(ctx, query)
	return int(f + 0.5), err
}

// scalar runs an instant query and returns the first sample's value, or 0 if
// the result set is empty (e.g. count() over no matching series).
func (c *Client) scalar(ctx context.Context, query string) (float64, error) {
	endpoint := c.base + "/api/v1/query?" + url.Values{"query": {query}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("victoriametrics: query %q: status %d", query, resp.StatusCode)
	}

	// Prometheus instant-query shape:
	//   {"status":"success","data":{"result":[{"value":[<ts>,"<float>"]}]}}
	var body struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value [2]json.RawMessage `json:"value"` // [ <unix ts>, "<stringified float>" ]
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("victoriametrics: decode: %w", err)
	}
	if body.Status != "success" {
		return 0, fmt.Errorf("victoriametrics: query %q: status %q", query, body.Status)
	}
	if len(body.Data.Result) == 0 {
		return 0, nil // empty vector → treat as zero
	}

	var raw string
	if err := json.Unmarshal(body.Data.Result[0].Value[1], &raw); err != nil {
		return 0, fmt.Errorf("victoriametrics: value: %w", err)
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("victoriametrics: parse %q: %w", raw, err)
	}
	return f, nil
}

// sample is one series from an instant vector query: its label set and value.
type sample struct {
	metric map[string]string
	value  float64
}

// vector runs an instant query and returns every series (labels + value),
// unlike scalar which keeps only the first value. Used for per-job `up`.
func (c *Client) vector(ctx context.Context, query string) ([]sample, error) {
	endpoint := c.base + "/api/v1/query?" + url.Values{"query": {query}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("victoriametrics: query %q: status %d", query, resp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string  `json:"metric"`
				Value  [2]json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("victoriametrics: decode: %w", err)
	}
	if body.Status != "success" {
		return nil, fmt.Errorf("victoriametrics: query %q: status %q", query, body.Status)
	}

	out := make([]sample, 0, len(body.Data.Result))
	for _, r := range body.Data.Result {
		var raw string
		if err := json.Unmarshal(r.Value[1], &raw); err != nil {
			return nil, fmt.Errorf("victoriametrics: value: %w", err)
		}
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("victoriametrics: parse %q: %w", raw, err)
		}
		out = append(out, sample{metric: r.Metric, value: f})
	}
	return out, nil
}
