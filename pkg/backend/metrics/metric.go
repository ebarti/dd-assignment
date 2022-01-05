package metrics

import (
	"bytes"
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/backend/errors"
	"strings"
)

type Metric interface {
	AddSample(*MetricSample) error // No implementations return an error here, but the interface is defined to allow for future implementations that do return an error
	Flush(timestamp int64) (*ComputedMetric, error)
}

type ComputedMetric struct {
	Name      string
	Timestamp int64
	Value     int64
	Groups    []*ComputedMetric
}

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

func (c *ComputedMetric) Render(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "Metric %s count: %d", c.Name, c.Value)
	for _, group := range c.Groups {
		fmt.Fprintf(buf, "\n\tWith tag name %s: %d", group.Name, group.Value)
		for _, subGroup := range group.Groups {
			fmt.Fprintf(buf, "\n\t\tWith tag name %s: %d", subGroup.Name, subGroup.Value)
		}
	}
}

func printIndented(buf *bytes.Buffer, indentation int, format string, args ...interface{}) {
	tabs := strings.Repeat("\t", indentation)
	buf.WriteString(fmt.Sprintf("\n%s%s", tabs, fmt.Sprintf(format, args...)))
}

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

type MetricSample struct {
	Name      string
	Tags      []*Tag
	Value     int64
	Timestamp int64
}

type Tag struct {
	Name  string
	Value string
}
