package metrics

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountMetric(t *testing.T) {
	flushtime := int64(105)
	tests := []struct {
		name    string
		samples []*MetricSample
		wantErr bool
		want    *ComputedMetric
	}{
		{
			name:    "test CountMetric return error when no samples",
			wantErr: true,
		},
		{
			name:    "test CountMetric without tags",
			samples: aMetricSamples,
			want:    aMetricSamplesComputed,
		},
		{
			name:    "test CountMetric with tags",
			samples: aTaggedMetricSamples,
			want:    aTaggedMetricSamplesComputed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCountMetric()
			for _, sample := range tt.samples {
				c.AddSample(sample)
			}
			got, err := c.Flush(flushtime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Equals(got))
			}
		})
	}
}

func TestCountSampling(t *testing.T) {
	tests := []struct {
		sampleValues []int64
		want         int64
	}{
		{
			sampleValues: nil,
			want:         0,
		},
		{
			sampleValues: []int64{12},
			want:         12,
		},
		{
			sampleValues: []int64{1, 2, 5, 0, 8, 3},
			want:         19,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("test count with sample values %v", tt.sampleValues)
		t.Run(name, func(t *testing.T) {
			c := NewCount()
			for _, sampleValue := range tt.sampleValues {
				c.AddSample(&MetricSample{Value: sampleValue})
			}
			assert.Equal(t, tt.want, c.Flush())
		})
	}
}
