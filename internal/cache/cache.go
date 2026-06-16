// Package cache provides a tiny stale-while-revalidate cache for a single value.
//
// It exists so page traffic never maps one-to-one onto upstream
// (VictoriaMetrics / Gitea) requests: the value is refreshed at most once per
// TTL, in the background, and callers are never blocked on an upstream call. On
// a refresh failure the last good value keeps being served, which gives the
// front-end graceful degradation for free.
package cache

import (
	"sync"
	"time"
)

// TTL caches one value of type T, refreshed by load at most once per ttl.
//
// It is safe for concurrent use.
type TTL[T any] struct {
	ttl  time.Duration
	load func() (T, error)

	mu         sync.Mutex
	val        T
	loaded     bool      // has load ever succeeded?
	expires    time.Time // when the current value becomes stale
	refreshing bool      // is a background refresh already in flight?
}

// New returns a cache that calls load to (re)populate its value.
//
// load owns its own timeout/context; it is intentionally not given a caller's
// request context, so a finished HTTP request can't cancel a shared refresh.
func New[T any](ttl time.Duration, load func() (T, error)) *TTL[T] {
	return &TTL[T]{ttl: ttl, load: load}
}

// Get returns the current value and whether one is available yet.
//
// It never blocks on load: if the value is missing or stale it kicks off a
// single background refresh and returns immediately. The cold-start call
// returns ok=false until the first refresh lands; pollers pick the value up on
// a later call. While an upstream is failing, the last good value keeps being
// returned (ok stays true).
func (c *TTL[T]) Get() (val T, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if (!c.loaded || time.Now().After(c.expires)) && !c.refreshing {
		c.refreshing = true
		go c.refresh()
	}
	return c.val, c.loaded
}

// refresh runs load once and stores the result. A failure is swallowed here so
// the previous value survives; the loader is expected to surface the failure
// via metrics instead. Leaving expires in the past means the next Get retries.
func (c *TTL[T]) refresh() {
	v, err := c.load()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.refreshing = false
	if err != nil {
		return
	}
	c.val = v
	c.loaded = true
	c.expires = time.Now().Add(c.ttl)
}
