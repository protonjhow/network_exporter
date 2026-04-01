package collector

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/syepes/network_exporter/monitor"
	"github.com/syepes/network_exporter/pkg/mtr"
	"github.com/syepes/network_exporter/pkg/ping"
	"github.com/syepes/network_exporter/pkg/tcp"

	"github.com/syepes/network_exporter/pkg/common"
)

// TestHistogramMetricsExposed verifies that histogram metrics appear in
// Prometheus exposition format alongside existing gauge metrics.
func TestHistogramMetricsExposed(t *testing.T) {
	reg := prometheus.NewRegistry()

	pingCollector := &PING{
		Monitor: &monitor.PING{},
		metrics: map[string]*ping.PingResult{
			"ping-target": {
				Success:  true,
				DestAddr: "example.com",
				DestIp:   "1.2.3.4",
				AvgTime:  5 * time.Millisecond,
				BestTime: 3 * time.Millisecond,
				WorstTime: 7 * time.Millisecond,
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
			"ping-target": {},
		},
	}

	tcpCollector := &TCP{
		Monitor: &monitor.TCPPort{},
		metrics: map[string]*tcp.TCPPortReturn{
			"tcp-target": {
				Success:  true,
				DestAddr: "example.com",
				DestIp:   "1.2.3.4",
				DestPort: "443",
				SrcIp:    "10.0.0.1",
				ConTime:  10 * time.Millisecond,
				Samples: []time.Duration{
					8 * time.Millisecond,
					10 * time.Millisecond,
					12 * time.Millisecond,
				},
				SampleCount:  3,
				SampleSumSec: 0.030,
			},
		},
		labels: map[string]map[string]string{
			"tcp-target": {},
		},
	}

	mtrCollector := &MTR{
		Monitor: &monitor.MTR{},
		metrics: map[string]*mtr.MtrResult{
			"mtr-target": {
				DestAddr: "example.com",
				Hops: []common.IcmpHop{
					{
						Success:     true,
						AddressFrom: "10.0.0.1",
						AddressTo:   "10.0.0.2",
						TTL:         1,
						Snt:         3,
						SntFail:     0,
						SumTime:     15 * time.Millisecond,
						AvgTime:     5 * time.Millisecond,
						BestTime:    3 * time.Millisecond,
						WorstTime:   7 * time.Millisecond,
						AllTime: []time.Duration{
							3 * time.Millisecond,
							5 * time.Millisecond,
							7 * time.Millisecond,
						},
					},
				},
			},
		},
		labels: map[string]map[string]string{
			"mtr-target": {},
		},
	}

	reg.MustRegister(pingCollector)
	reg.MustRegister(tcpCollector)
	reg.MustRegister(mtrCollector)

	// Gather all metrics
	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Render to exposition format
	var sb strings.Builder
	enc := expfmt.NewEncoder(&sb, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range metricFamilies {
		if err := enc.Encode(mf); err != nil {
			t.Fatalf("failed to encode metric family: %v", err)
		}
	}
	output := sb.String()

	// Verify histogram metrics exist
	histogramMetrics := []string{
		"ping_rtt_duration_seconds_bucket",
		"ping_rtt_duration_seconds_count",
		"ping_rtt_duration_seconds_sum",
		"tcp_connection_duration_seconds_bucket",
		"tcp_connection_duration_seconds_count",
		"tcp_connection_duration_seconds_sum",
		"mtr_rtt_duration_seconds_bucket",
		"mtr_rtt_duration_seconds_count",
		"mtr_rtt_duration_seconds_sum",
	}

	for _, metric := range histogramMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("expected metric %q in output, not found", metric)
		}
	}

	// Verify existing gauge metrics still present
	gaugeMetrics := []string{
		"ping_rtt_seconds",
		"ping_status",
		"ping_loss_percent",
		"tcp_connection_seconds",
		"tcp_connection_status",
		"mtr_rtt_seconds",
		"mtr_hops",
	}

	for _, metric := range gaugeMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("expected gauge metric %q in output, not found", metric)
		}
	}

	// Verify bucket boundaries appear (spot check a few)
	bucketChecks := []string{
		`le="0.001"`,  // 1ms bucket
		`le="0.01"`,   // 10ms bucket
		`le="0.1"`,    // 100ms bucket
		`le="+Inf"`,   // +Inf bucket always present
	}

	for _, check := range bucketChecks {
		if !strings.Contains(output, check) {
			t.Errorf("expected bucket boundary %q in output, not found", check)
		}
	}
}
