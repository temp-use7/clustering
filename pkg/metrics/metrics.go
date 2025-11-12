package metrics

import (
	"fmt"
	"sort"
	"sync"
)

var (
	countersMu sync.Mutex
	counters   = map[string]float64{}
	gaugesMu   sync.Mutex
	gauges     = map[string]float64{}
)

func IncCounter(name string) {
	countersMu.Lock()
	counters[name] += 1
	countersMu.Unlock()
}

func AddCounter(name string, delta float64) {
	countersMu.Lock()
	counters[name] += delta
	countersMu.Unlock()
}

func SetGauge(name string, value float64) {
	gaugesMu.Lock()
	gauges[name] = value
	gaugesMu.Unlock()
}

// RenderPrometheus returns metrics in prometheus text exposition format.
func RenderPrometheus() string {
	lines := []string{}
	countersMu.Lock()
	for k, v := range counters {
		lines = append(lines, fmt.Sprintf("%s %g", sanitize(k), v))
	}
	countersMu.Unlock()
	gaugesMu.Lock()
	for k, v := range gauges {
		lines = append(lines, fmt.Sprintf("%s %g", sanitize(k), v))
	}
	gaugesMu.Unlock()
	sort.Strings(lines)
	return joinLines(lines)
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == ':' {
			out = append(out, r)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	n := 0
	for _, l := range lines {
		n += len(l) + 1
	}
	b := make([]byte, 0, n)
	for i, l := range lines {
		b = append(b, l...)
		if i != len(lines)-1 {
			b = append(b, '\n')
		}
	}
	return string(b)
}

