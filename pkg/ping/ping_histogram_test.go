package ping

import (
	"testing"
	"time"
)

func TestPingResultContainsAllSamples(t *testing.T) {
	result := PingResult{
		AllTime: []time.Duration{
			1 * time.Millisecond,
			2 * time.Millisecond,
			3 * time.Millisecond,
		},
	}
	if len(result.AllTime) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(result.AllTime))
	}
}
