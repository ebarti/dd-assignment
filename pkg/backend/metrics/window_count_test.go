package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWindowCountMetric(t *testing.T) {
	tests := []struct {
		name    string
		window  int64
		samples []*MetricSample
		wantErr bool
		want    *ComputedMetric
	}{
		{
			name:    "test WindowMetric return error when no samples",
			wantErr: true,
		},
		{
			name:    "test WindowMetric without tags",
			window:  5,
			samples: aMetricSamples,
			want:    aMetricSamplesComputed,
		},
		{
			name:    "test WindowMetric with tags",
			window:  5,
			samples: aTaggedMetricSamples,
			want:    aTaggedMetricSamplesComputed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewWindowCountMetric(tt.window)
			for _, sample := range tt.samples {
				c.AddSample(sample)
			}
			got, err := c.Flush(aFlushTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Equals(got))
			}
		})
	}
}

func TestWindowCountSampling(t *testing.T) {
	windowMetricFlushTime := int64(105)

	tests := []struct {
		name   string
		window int64
		want   int64
	}{
		{
			name:   "all samples in window",
			window: int64(5),
			want:   int64(4),
		},
		{
			name:   "some samples in window",
			window: int64(3),
			want:   int64(2),
		},
		{
			name:   "no samples in window",
			window: int64(1),
			want:   int64(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewWindowCount(tt.window)
			for _, sample := range aTaggedMetricSamples {
				c.AddSample(sample)
			}
			assert.Equal(t, tt.want, c.Flush(windowMetricFlushTime))
		})
	}
}
