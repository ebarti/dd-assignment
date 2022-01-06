package logs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogMeasure_Measure(t *testing.T) {
	tests := []struct {
		measure      string
		want         int64
		isMeasurable bool
	}{
		{
			measure:      "status",
			want:         200,
			isMeasurable: true,
		},
		{
			measure:      "aMeasurableAttribute",
			want:         1,
			isMeasurable: true,
		},
		{
			measure:      "nested.aMeasurableAttribute",
			want:         2,
			isMeasurable: true,
		},

		{
			measure:      "nested.nested.aMeasurableAttribute",
			want:         3,
			isMeasurable: true,
		},
		{
			measure:      "nested.nested.nested.aMeasurableAttribute",
			want:         4,
			isMeasurable: true,
		},
		// Non-measurable attributes
		{
			measure:      "host",
			isMeasurable: false,
		},
		{
			measure:      "service",
			isMeasurable: false,
		},
		{
			measure:      "message",
			isMeasurable: false,
		},
		// Non-existent attributes:
		{
			measure:      "aNonExistentAttribute",
			isMeasurable: false,
		},
		{
			measure:      "nested.aNonExistentAttribute",
			isMeasurable: false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("measure %s attribute", tt.measure), func(t *testing.T) {
			d := &LogMeasure{
				name: tt.measure,
			}
			if tt.isMeasurable {
				assert.Equal(t, tt.want, *d.Measure(aProcessedLog))
			} else {
				assert.Nil(t, d.Measure(aProcessedLog))
			}
		})
	}
}
