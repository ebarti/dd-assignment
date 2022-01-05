package metrics

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var computedMetricSample = &ComputedMetric{
	Name:      "test",
	Timestamp: 1641308532,
	Value:     12,
	Groups: []*ComputedMetric{
		{
			Name:      "host",
			Timestamp: 1641308532,
			Value:     12,
			Groups: []*ComputedMetric{
				{
					Name:      "host1",
					Timestamp: 1641308532,
					Value:     6,
				},
				{
					Name:      "host2",
					Timestamp: 1641308532,
					Value:     4,
				},
				{
					Name:      "host3",
					Timestamp: 1641308532,
					Value:     2,
				},
			},
		},
		{
			Name:      "http.path.section",
			Timestamp: 1641308532,
			Value:     8,
			Groups: []*ComputedMetric{
				{
					Name:      "section1",
					Timestamp: 1641308532,
					Value:     5,
				},
				{
					Name:      "section2",
					Timestamp: 1641308532,
					Value:     3,
				},
			},
		},
	},
}

func TestComputedMetric_GetValue(t *testing.T) {
	tests := []struct {
		tag     *Tag
		want    int64
		wantErr bool
	}{
		{
			tag:  nil,
			want: 12,
		},
		{
			tag:  &Tag{Name: "http.path.section"},
			want: 8,
		},
		{
			tag:  &Tag{Name: "host", Value: "host1"},
			want: 6,
		},
		{
			tag:     &Tag{Name: "nonexistent"},
			wantErr: true,
		},

		{
			tag:     &Tag{Name: "host", Value: "nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		name := "computedMetric.GetValue"
		if tt.tag != nil {
			name += fmt.Sprintf(" for tag %s:%s", tt.tag.Name, tt.tag.Value)
		}
		t.Run(name, func(t *testing.T) {
			got, err := computedMetricSample.GetValue(tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestComputedMetric_Equals(t *testing.T) {
	computedMetricSample1 := &ComputedMetric{
		Name:      "test",
		Timestamp: 1641308532,
		Value:     12,
		Groups: []*ComputedMetric{
			{
				Name:      "host",
				Timestamp: 1641308532,
				Value:     12,
				Groups: []*ComputedMetric{
					{
						Name:      "host1",
						Timestamp: 1641308532,
						Value:     6,
					},
					{
						Name:      "host2",
						Timestamp: 1641308532,
						Value:     4,
					},
					{
						Name:      "host3",
						Timestamp: 1641308532,
						Value:     2,
					},
				},
			},
			{
				Name:      "http.path.section",
				Timestamp: 1641308532,
				Value:     8,
				Groups: []*ComputedMetric{
					{
						Name:      "section1",
						Timestamp: 1641308532,
						Value:     5,
					},
					{
						Name:      "section2",
						Timestamp: 1641308532,
						Value:     3,
					},
				},
			},
		},
	}
	computedMetricSample2 := &ComputedMetric{
		Name:      "test",
		Timestamp: 1641308532,
		Value:     12,
		Groups: []*ComputedMetric{
			{
				Name:      "http.path.section",
				Timestamp: 1641308532,
				Value:     8,
				Groups: []*ComputedMetric{
					{
						Name:      "section1",
						Timestamp: 1641308532,
						Value:     5,
					},
					{
						Name:      "section2",
						Timestamp: 1641308532,
						Value:     3,
					},
				},
			},
			{
				Name:      "host",
				Timestamp: 1641308532,
				Value:     12,
				Groups: []*ComputedMetric{
					{
						Name:      "host2",
						Timestamp: 1641308532,
						Value:     4,
					},
					{
						Name:      "host3",
						Timestamp: 1641308532,
						Value:     2,
					},
					{
						Name:      "host1",
						Timestamp: 1641308532,
						Value:     6,
					},
				},
			},
		},
	}
	assert.True(t, computedMetricSample1.Equals(computedMetricSample2))

	// verify changing a value makes the metrics different
	prevVal := computedMetricSample2.Groups[1].Groups[2].Value
	computedMetricSample2.Groups[1].Groups[2].Value = 0
	assert.False(t, computedMetricSample1.Equals(computedMetricSample2))

	// revert the change
	computedMetricSample2.Groups[1].Groups[2].Value = prevVal
	assert.True(t, computedMetricSample1.Equals(computedMetricSample2))

	// add an extra group to the other compared sample
	idxToDelete := len(computedMetricSample2.Groups[1].Groups)
	computedMetricSample2.Groups[1].Groups = append(computedMetricSample2.Groups[1].Groups, &ComputedMetric{Name: "test", Value: 1})
	assert.False(t, computedMetricSample1.Equals(computedMetricSample2))

	// revert the change
	computedMetricSample2.Groups[1].Groups = computedMetricSample2.Groups[1].Groups[:idxToDelete]
	assert.True(t, computedMetricSample1.Equals(computedMetricSample2))

	// add an extra group to the compared sample
	computedMetricSample1.Groups[1].Groups = append(computedMetricSample2.Groups[1].Groups, &ComputedMetric{Name: "test", Value: 1})
	assert.False(t, computedMetricSample1.Equals(computedMetricSample2))
}

func TestComputedMetric_Render(t *testing.T) {
	var buf bytes.Buffer
	computedMetricSample.Render(&buf)
	expected := "Metric test count: 12\n\tWith tag name host: 12\n\t\tWith tag name host1: 6\n\t\tWith tag name host2: 4\n\t\tWith tag name host3: 2\n\tWith tag name http.path.section: 8\n\t\tWith tag name section1: 5\n\t\tWith tag name section2: 3"
	assert.Equal(t, len([]byte(expected)), len(buf.Bytes()))
	assert.Equal(t, expected, buf.String())
}
