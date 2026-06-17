package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitFor polls cond until it is true or a generous deadline passes. The cache
// refreshes in the background, so tests synchronise on observable state rather
// than on fixed sleeps.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condition not met within deadline")
}

// TestColdStart confirms the first Get reports "no value yet" while kicking off
// the background load, and the value appears on a later Get.
func TestColdStart(t *testing.T) {
	c := New(time.Minute, func() (int, error) { return 42, nil })

	if _, ok := c.Get(); ok {
		t.Fatal("cold Get should report ok=false")
	}
	var val int
	waitFor(t, func() bool {
		v, ok := c.Get()
		val = v
		return ok
	})
	if val != 42 {
		t.Errorf("val = %d, want 42", val)
	}
}

// TestStaleWhileRevalidate is the core contract: once a value is loaded, a later
// refresh failure must not drop it — the last good value keeps being served.
func TestStaleWhileRevalidate(t *testing.T) {
	var fail atomic.Bool
	var calls atomic.Int64
	c := New(20*time.Millisecond, func() (int, error) {
		calls.Add(1)
		if fail.Load() {
			return -1, errors.New("upstream down")
		}
		return 7, nil
	})

	// Prime the cache with a good value.
	waitFor(t, func() bool { _, ok := c.Get(); return ok })
	if v, _ := c.Get(); v != 7 {
		t.Fatalf("primed value = %d, want 7", v)
	}

	// Let it go stale, then break the loader.
	time.Sleep(40 * time.Millisecond)
	fail.Store(true)

	// The Get that finds it stale triggers a background refresh that fails;
	// the previous value must survive (ok stays true).
	if v, ok := c.Get(); !ok || v != 7 {
		t.Fatalf("stale Get = (%d,%v), want (7,true)", v, ok)
	}
	waitFor(t, func() bool { return calls.Load() >= 2 }) // the failing refresh ran
	if v, ok := c.Get(); !ok || v != 7 {
		t.Errorf("after failed refresh = (%d,%v), want last good (7,true)", v, ok)
	}
}

// TestSingleFlight ensures concurrent Gets over a missing value start exactly
// one background load, not one per caller.
func TestSingleFlight(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int64
	c := New(time.Minute, func() (int, error) {
		calls.Add(1)
		<-release // hold the single refresh open while callers pile up
		return 1, nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Get() }()
	}
	wg.Wait()

	waitFor(t, func() bool { return calls.Load() >= 1 }) // the one refresh started
	if got := calls.Load(); got != 1 {
		t.Errorf("loader called %d times, want 1 (single-flight)", got)
	}
	close(release)
}
