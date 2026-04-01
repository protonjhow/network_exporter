package config

import (
	"sort"
	"testing"
)

func TestHistogramBucketsValidation(t *testing.T) {
	tests := []struct {
		name    string
		buckets []float64
		wantErr bool
	}{
		{
			name:    "valid sorted buckets",
			buckets: []float64{0.001, 0.01, 0.1, 1.0},
			wantErr: false,
		},
		{
			name:    "valid unsorted buckets get sorted",
			buckets: []float64{1.0, 0.001, 0.1, 0.01},
			wantErr: false,
		},
		{
			name:    "negative bucket rejected",
			buckets: []float64{-0.1, 0.001, 0.01},
			wantErr: true,
		},
		{
			name:    "zero bucket rejected",
			buckets: []float64{0, 0.001, 0.01},
			wantErr: true,
		},
		{
			name:    "duplicate bucket rejected",
			buckets: []float64{0.001, 0.001, 0.01},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHistogramBuckets(tt.buckets)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestHistogramBucketsAutoSort(t *testing.T) {
	buckets := []float64{1.0, 0.001, 0.5, 0.01}
	sort.Float64s(buckets)

	for i := 1; i < len(buckets); i++ {
		if buckets[i] <= buckets[i-1] {
			t.Errorf("buckets not sorted: %v", buckets)
		}
	}

	expected := []float64{0.001, 0.01, 0.5, 1.0}
	for i, b := range buckets {
		if b != expected[i] {
			t.Errorf("index %d: expected %f, got %f", i, expected[i], b)
		}
	}
}
