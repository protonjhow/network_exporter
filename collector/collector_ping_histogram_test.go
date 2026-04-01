package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/syepes/network_exporter/monitor"
	"github.com/syepes/network_exporter/pkg/ping"
)

func TestPingHistogramEmitted(t *testing.T) {
	p := &PING{
		Monitor: &monitor.PING{},
		metrics: map[string]*ping.PingResult{
			"test-target": {
				Success:  true,
				DestAddr: "example.com",
				DestIp:   "1.2.3.4",
				AvgTime:  5 * time.Millisecond,
				SumTime:  15 * time.Millisecond,
				AllTime: []time.Duration{
					3 * time.Millisecond,
					5 * time.Millisecond,
					7 * time.Millisecond,
				},
				SntSummary:     3,
				SntFailSummary: 0,
			},
		},
		labels: map[string]map[string]string{
			"test-target": {},
		},
	}

	ch := make(chan prometheus.Metric, 50)
	p.Collect(ch)
	close(ch)

	var histogramFound bool
	for m := range ch {
		d := &dto.Metric{}
		if err := m.Write(d); err != nil {
			continue
		}
		if d.Histogram != nil {
			histogramFound = true
			if d.Histogram.GetSampleCount() != 3 {
				t.Errorf("expected sample count 3, got %d", d.Histogram.GetSampleCount())
			}
			if d.Histogram.GetSampleSum() == 0 {
				t.Error("expected non-zero sample sum")
			}
			if len(d.Histogram.GetBucket()) == 0 {
				t.Error("expected non-empty buckets")
			}
		}
	}

	if !histogramFound {
		t.Fatal("no histogram metric found in PING collector output")
	}
}

func TestComputeBucketCounts(t *testing.T) {
	samples := []time.Duration{
		1 * time.Millisecond,   // 0.001s
		5 * time.Millisecond,   // 0.005s
		50 * time.Millisecond,  // 0.05s
		500 * time.Millisecond, // 0.5s
	}
	buckets := []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}

	counts := computeBucketCounts(samples, buckets)

	expected := map[float64]uint64{
		0.001: 1, // 1ms
		0.005: 2, // 1ms, 5ms
		0.01:  2,
		0.05:  3, // 1ms, 5ms, 50ms
		0.1:   3,
		0.5:   4, // all
		1.0:   4,
	}

	for b, want := range expected {
		got := counts[b]
		if got != want {
			t.Errorf("bucket %.3f: expected %d, got %d", b, want, got)
		}
	}
}
