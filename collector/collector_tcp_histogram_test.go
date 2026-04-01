package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/syepes/network_exporter/monitor"
	"github.com/syepes/network_exporter/pkg/tcp"
)

func TestTCPHistogramEmitted(t *testing.T) {
	p := &TCP{
		Monitor: &monitor.TCPPort{},
		metrics: map[string]*tcp.TCPPortReturn{
			"test-target": {
				Success:  true,
				DestAddr: "example.com",
				DestIp:   "1.2.3.4",
				DestPort: "443",
				SrcIp:    "10.0.0.1",
				ConTime:  5 * time.Millisecond,
				Samples: []time.Duration{
					3 * time.Millisecond,
					5 * time.Millisecond,
					7 * time.Millisecond,
				},
				SampleCount:  3,
				SampleSumSec: 0.015,
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
		}
	}

	if !histogramFound {
		t.Fatal("no histogram metric found in TCP collector output")
	}
}
