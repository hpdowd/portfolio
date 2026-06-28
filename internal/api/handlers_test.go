package api

import (
	"testing"
	"time"

	"git.henrydowd.dev/henry/portfolio/internal/gitea"
	"git.henrydowd.dev/henry/portfolio/internal/vm"
)

// fakeReporter records the metrics interactions New performs, so we can assert
// the upstreams are pre-registered without standing up a real metrics registry.
type fakeReporter struct {
	registered []string
}

func (f *fakeReporter) SetUpstreamUp(string, bool) {}
func (f *fakeReporter) RegisterUpstreams(names ...string) {
	f.registered = append(f.registered, names...)
}

// TestNewRegistersUpstreams locks in the fix for absent()-based alerts: New must
// pre-register every upstream at construction, so the series exist for scraping
// before any /api traffic exercises a loader. The cache is lazy, so no upstream
// call happens here.
func TestNewRegistersUpstreams(t *testing.T) {
	var r fakeReporter
	vmClient := vm.New("http://vm.invalid", time.Second)
	giteaClient := gitea.New("http://gitea.invalid", time.Second)

	_ = New("v", vmClient, giteaClient, "owner/repo", 6, time.Minute, time.Second, &r)

	got := map[string]bool{}
	for _, n := range r.registered {
		got[n] = true
	}
	for _, want := range []string{upstreamVM, upstreamGitea} {
		if !got[want] {
			t.Errorf("New did not register upstream %q (registered: %v)", want, r.registered)
		}
	}
}
