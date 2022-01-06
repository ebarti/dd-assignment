package metrics

import (
	"bytes"
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/errors"
)

// Metric defines an interface that a metric must implement.
type Metric interface {
	AddSample(*MetricSample) error // No implementations return an error here, but the interface is defined to allow for future implementations that do return an error
	Flush(timestamp int64) (*ComputedMetric, error)
}

// ComputedMetric is a struct that contains the computed metric, generated after a flush.
type ComputedMetric struct {
	Name      string
	Timestamp int64
	Value     int64
	Groups    []*ComputedMetric
}

// Equals returns true if the two computed metrics are equal.
func (c *ComputedMetric) Equals(other *ComputedMetric) bool {
	if c.Name != other.Name {
		return false
	}
	if c.Timestamp != other.Timestamp {
		return false
	}
	if c.Value != other.Value {
		return false
	}
	if len(c.Groups) != len(other.Groups) {
		return false
	}

	for _, group := range c.Groups {
		found := false
		for _, otherGroup := range other.Groups {
			if group.Equals(otherGroup) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Render writes the string representation of the computed metric on the provided buffer.
func (c *ComputedMetric) Render(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "Metric %s count: %d", c.Name, c.Value)
	for _, group := range c.Groups {
		fmt.Fprintf(buf, "\n\tWith tag name %s: %d", group.Name, group.Value)
		for _, subGroup := range group.Groups {
			fmt.Fprintf(buf, "\n\t\tWith tag name %s: %d", subGroup.Name, subGroup.Value)
		}
	}
}

// GetValue returns the value of the computed metric.
func (c *ComputedMetric) GetValue(tag *Tag) (int64, error) {
	if tag == nil {
		return c.Value, nil
	}
	for _, group := range c.Groups {
		if group.Name == tag.Name {
			if tag.Value == "" {
				return group.Value, nil
			}
			for _, tagGroup := range group.Groups {
				if tagGroup.Name == tag.Value {
					return tagGroup.Value, nil
				}
			}
		}
	}
	return 0, errors.NewCouldNotComputeMetricForTagError(tag.Name, tag.Value)
}

// MetricSample represents a metric extracted from a log.
type MetricSample struct {
	Name      string
	Tags      []*Tag
	Value     int64
	Timestamp int64
}

// Tag represents a tag name and its value for a given log.
type Tag struct {
	Name  string
	Value string
}
