package collector

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/syepes/network_exporter/monitor"
	"github.com/syepes/network_exporter/pkg/common"
	"github.com/syepes/network_exporter/pkg/mtr"
	"github.com/syepes/network_exporter/pkg/ping"
)

// TestPingHistogramNegativeSuccessCount verifies that when SntFailSummary
// exceeds SntSummary (corrupted data), the collector does not panic from
// a uint64 underflow and instead clamps to 0.
func TestPingHistogramNegativeSuccessCount(t *testing.T) {
	p := &PING{
		Monitor: &monitor.PING{},
		metrics: map[string]*ping.PingResult{
			"bad-target": {
				Success:        true,
				DestAddr:       "example.com",
				DestIp:         "1.2.3.4",
				SumTime:        5 * time.Millisecond,
				AllTime:        []time.Duration{5 * time.Millisecond},
				SntSummary:     3,
				SntFailSummary: 5, // More failures than sent -- corrupted
			},
		},
		labels: map[string]map[string]string{
			"bad-target": {},
		},
	}

	ch := make(chan prometheus.Metric, 50)
	// Must not panic
	p.Collect(ch)
	close(ch)

	for m := range ch {
		d := &dto.Metric{}
		if err := m.Write(d); err != nil {
			continue
		}
		if d.Histogram != nil {
			if d.Histogram.GetSampleCount() != 0 {
				t.Errorf("expected clamped sample count 0, got %d", d.Histogram.GetSampleCount())
			}
		}
	}
}

// TestMTRHistogramNegativeSuccessCount verifies the same uint64 underflow
// guard for MTR per-hop histograms.
func TestMTRHistogramNegativeSuccessCount(t *testing.T) {
	p := &MTR{
		Monitor: &monitor.MTR{},
		metrics: map[string]*mtr.MtrResult{
			"bad-target": {
				DestAddr: "example.com",
				Hops: []common.IcmpHop{
					{
						Success:     true,
						AddressFrom: "10.0.0.1",
						AddressTo:   "10.0.0.2",
						TTL:         1,
						Snt:         2,
						SntFail:     5, // More failures than sent
						SumTime:     5 * time.Millisecond,
						AllTime:     []time.Duration{5 * time.Millisecond},
					},
				},
			},
		},
		labels: map[string]map[string]string{
			"bad-target": {},
		},
	}

	ch := make(chan prometheus.Metric, 50)
	p.Collect(ch)
	close(ch)

	for m := range ch {
		d := &dto.Metric{}
		if err := m.Write(d); err != nil {
			continue
		}
		if d.Histogram != nil {
			if d.Histogram.GetSampleCount() != 0 {
				t.Errorf("expected clamped sample count 0, got %d", d.Histogram.GetSampleCount())
			}
		}
	}
}

// TestPingNoHistogramOnAllProbesFailed verifies that no histogram is emitted
// when all probes fail (empty AllTime slice).
func TestPingNoHistogramOnAllProbesFailed(t *testing.T) {
	p := &PING{
		Monitor: &monitor.PING{},
		metrics: map[string]*ping.PingResult{
			"failed-target": {
				Success:        false,
				DestAddr:       "example.com",
				DestIp:         "1.2.3.4",
				AllTime:        nil,
				SntSummary:     10,
				SntFailSummary: 10,
			},
		},
		labels: map[string]map[string]string{
			"failed-target": {},
		},
	}

	ch := make(chan prometheus.Metric, 50)
	p.Collect(ch)
	close(ch)

	for m := range ch {
		d := &dto.Metric{}
		if err := m.Write(d); err != nil {
			continue
		}
		if d.Histogram != nil {
			t.Fatal("should not emit histogram when all probes failed")
		}
	}
}
