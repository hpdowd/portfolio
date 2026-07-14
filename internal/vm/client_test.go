package vm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at a throwaway server running handler.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(srv.URL, 2*time.Second)
}

// TestScalar covers the instant-query value parser, including the edge cases
// that matter: an empty result vector is zero (not an error), and malformed or
// failed responses surface as errors so the cache keeps the last good value.
func TestScalar(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		want    float64
		wantErr bool
	}{
		{"integer", `{"status":"success","data":{"result":[{"value":[1,"3"]}]}}`, 200, 3, false},
		{"float", `{"status":"success","data":{"result":[{"value":[1,"0.0633"]}]}}`, 200, 0.0633, false},
		{"empty vector is zero", `{"status":"success","data":{"result":[]}}`, 200, 0, false},
		{"http error", `{}`, 500, 0, true},
		{"status not success", `{"status":"error","data":{"result":[]}}`, 200, 0, true},
		{"value not a string", `{"status":"success","data":{"result":[{"value":[1,123]}]}}`, 200, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				if tt.status != 200 {
					w.WriteHeader(tt.status)
				}
				_, _ = w.Write([]byte(tt.body))
			})
			got, err := c.scalar(context.Background(), "count(up)")
			if (err != nil) != tt.wantErr {
				t.Fatalf("scalar err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("scalar = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestScalarIntRounding confirms scalarInt rounds to nearest rather than
// truncating (count() results are integers, but a rate could be fractional).
func TestScalarIntRounding(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"success","data":{"result":[{"value":[1,"2.6"]}]}}`)
	})
	got, err := c.scalarInt(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Errorf("scalarInt(2.6) = %d, want 3", got)
	}
}

// TestFetch checks that each fixed query lands in the right Snapshot field.
func TestFetch(t *testing.T) {
	values := map[string]string{
		"count(up == 1)": "22",
		"count(up)":      "22",
		`count(ALERTS{alertstate="firing",severity=~"warning|critical",alertname!~"KubeCPUOvercommit|KubeMemoryOvercommit"})`: "3",
		`sum(rate(portfolio_http_requests_total[5m]))`: "0.0633",
		"min by (job) (up)":                            "1", // no job label in the mock → curated to nil
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("query")
		v, ok := values[q]
		if !ok {
			t.Errorf("unexpected query: %q", q)
			v = "0"
		}
		_, _ = fmt.Fprintf(w, `{"status":"success","data":{"result":[{"value":[1,%q]}]}}`, v)
	})
	got, err := c.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	want := Snapshot{TargetsUp: 22, TargetsTotal: 22, AlertsFiring: 3, RequestRate: 0.0633}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Fetch = %+v, want %+v", got, want)
	}
}
