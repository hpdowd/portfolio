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
	return s, nil
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
