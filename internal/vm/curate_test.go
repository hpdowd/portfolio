package vm

import "testing"

func TestCurateServices(t *testing.T) {
	// Raw job labels as VictoriaMetrics might report them, mixed up/down, plus
	// jobs that map to no diagram node (they must be dropped).
	jobs := map[string]bool{
		"portfolio":                          true,
		"longhorn-backend-manager":           true,
		"vmsingle-vm-victoria-metrics-stack": true,
		"vmalert":                            false, // one victoria job down -> canonical down
		"grafana":                            true,
		"node-exporter":                      true, // not a diagram node -> dropped
		"kubelet":                            true, // dropped
	}
	got := curateServices(jobs)

	want := map[string]bool{
		"portfolio":       true,
		"longhorn":        true,
		"victoriametrics": false, // AND over vmsingle(up) + vmalert(down)
		"grafana":         true,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("service %q: got %v, want %v", k, got[k], v)
		}
	}
	if _, ok := got["node-exporter"]; ok {
		t.Errorf("unmapped job leaked into output: %v", got)
	}
}

func TestCurateServicesEmpty(t *testing.T) {
	if got := curateServices(map[string]bool{"kubelet": true}); got != nil {
		t.Errorf("expected nil when nothing maps, got %v", got)
	}
}
