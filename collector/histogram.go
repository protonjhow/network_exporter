package collector

import "time"

// defaultHistogramBuckets returns bucket boundaries tuned for network latency.
func defaultHistogramBuckets() []float64 {
	return []float64{
		0.0001,  // 100us
		0.00025, // 250us
		0.0005,  // 500us
		0.001,   // 1ms
		0.0025,  // 2.5ms
		0.005,   // 5ms
		0.01,    // 10ms
		0.025,   // 25ms
		0.05,    // 50ms
		0.1,     // 100ms
		0.25,    // 250ms
		0.5,     // 500ms
		1.0,     // 1s
	}
}

// computeBucketCounts sorts samples into cumulative histogram buckets.
func computeBucketCounts(samples []time.Duration, buckets []float64) map[float64]uint64 {
	counts := make(map[float64]uint64, len(buckets))
	for _, b := range buckets {
		counts[b] = 0
	}
	for _, s := range samples {
		sec := s.Seconds()
		for _, b := range buckets {
			if sec <= b {
				counts[b]++
			}
		}
	}
	return counts
}
