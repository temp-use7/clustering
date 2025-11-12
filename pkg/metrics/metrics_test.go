package metrics

import "testing"

func TestRenderPrometheus(t *testing.T) {
	IncCounter("test_counter_total")
	SetGauge("test_gauge", 42)
	out := RenderPrometheus()
	if out == "" || !contains(out, "test_counter_total") || !contains(out, "test_gauge") {
		t.Fatalf("unexpected metrics: %q", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (index(s, sub) >= 0)))
}

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

