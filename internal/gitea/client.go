// Package gitea reads PUBLIC repository data from a Gitea instance via its REST
// API, anonymously (no token).
//
// The target instance has REQUIRE_SIGNIN_VIEW disabled, so public repositories
// are readable without authentication. Only a curated subset of each commit is
// returned, and messages are reduced to their first line — the service never
// echoes raw repository internals to the page.
package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client reads from a Gitea instance at base.
type Client struct {
	base string
	http *http.Client
}

// New returns a Client that talks to the Gitea instance at base, applying
// timeout to each request.
func New(base string, timeout time.Duration) *Client {
	return &Client{base: base, http: &http.Client{Timeout: timeout}}
}

// Commit is the curated view of a repository commit shown on the activity panel.
type Commit struct {
	SHA     string    `json:"sha"`     // short hash
	Summary string    `json:"summary"` // first line of the commit message only
	Author  string    `json:"author"`
	URL     string    `json:"url"`
	When    time.Time `json:"when"`
}

// RecentCommits returns up to limit commits from the default branch of
// "owner/repo". The repository must be public — this client sends no credentials.
func (c *Client) RecentCommits(ctx context.Context, ownerRepo string, limit int) ([]Commit, error) {
	owner, repo, ok := strings.Cut(ownerRepo, "/")
	if !ok {
		return nil, fmt.Errorf("gitea: repo must be \"owner/name\", got %q", ownerRepo)
	}

	// stat/verification off keeps the payload small — we only want the basics.
	q := url.Values{
		"limit":        {fmt.Sprint(limit)},
		"stat":         {"false"},
		"verification": {"false"},
	}
	endpoint := fmt.Sprintf("%s/api/v1/repos/%s/%s/commits?%s",
		c.base, url.PathEscape(owner), url.PathEscape(repo), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitea: commits %q: status %d", ownerRepo, resp.StatusCode)
	}

	// Only the fields we surface are decoded; the rest of Gitea's response is ignored.
	var raw []struct {
		SHA     string `json:"sha"`
		HTMLURL string `json:"html_url"`
		Commit  struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("gitea: decode: %w", err)
	}

	commits := make([]Commit, 0, len(raw))
	for _, r := range raw {
		summary, _, _ := strings.Cut(r.Commit.Message, "\n") // first line only
		sha := r.SHA
		if len(sha) > 8 {
			sha = sha[:8]
		}
		commits = append(commits, Commit{
			SHA:     sha,
			Summary: strings.TrimSpace(summary),
			Author:  r.Commit.Author.Name,
			URL:     r.HTMLURL,
			When:    r.Commit.Author.Date,
		})
	}
	return commits, nil
}
