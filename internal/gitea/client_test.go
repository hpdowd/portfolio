package gitea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestRecentCommits verifies the curation contract: the SHA is truncated to 8
// chars, the summary is the first line of the message only (trimmed), and the
// request hits the expected path with the limit propagated.
func TestRecentCommits(t *testing.T) {
	const body = `[
	  {"sha":"63370eac8324e1a9110c76017f092e64f018b3c0","html_url":"https://git.example/c/63370eac","commit":{"message":"portfolio: correct GHCR owner\n\nlong body that must not leak","author":{"name":"Henry Dowd","date":"2026-06-17T03:57:24+01:00"}}},
	  {"sha":"short","html_url":"https://git.example/c/short","commit":{"message":"  spaced summary  ","author":{"name":"X","date":"2026-06-17T03:00:00+01:00"}}}
	]`
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL, 2*time.Second)

	commits, err := c.RecentCommits(context.Background(), "henry/homelab", 6)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}

	// First commit: long SHA truncated, message reduced to its first line.
	if commits[0].SHA != "63370eac" {
		t.Errorf("SHA = %q, want 63370eac", commits[0].SHA)
	}
	if commits[0].Summary != "portfolio: correct GHCR owner" {
		t.Errorf("Summary = %q (the body must not leak)", commits[0].Summary)
	}
	if commits[0].Author != "Henry Dowd" {
		t.Errorf("Author = %q", commits[0].Author)
	}
	if commits[0].URL != "https://git.example/c/63370eac" {
		t.Errorf("URL = %q", commits[0].URL)
	}

	// Second commit: a short SHA is left intact; the summary is trimmed.
	if commits[1].SHA != "short" {
		t.Errorf("short SHA = %q, want intact", commits[1].SHA)
	}
	if commits[1].Summary != "spaced summary" {
		t.Errorf("trimmed summary = %q", commits[1].Summary)
	}

	if gotPath != "/api/v1/repos/henry/homelab/commits" {
		t.Errorf("request path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "limit=6") {
		t.Errorf("query = %q, want limit=6", gotQuery)
	}
}

// TestRecentCommitsBadRepo rejects a repo spec that isn't "owner/name" before
// any network call is made.
func TestRecentCommitsBadRepo(t *testing.T) {
	c := New("http://unused.invalid", time.Second)
	if _, err := c.RecentCommits(context.Background(), "no-slash", 5); err == nil {
		t.Fatal("expected an error for a repo spec without a slash")
	}
}
