package metrics

import (
	"github.com/ebarti/dd-assignment/pkg/backend/errors"
	"sync/atomic" // Atomic operations are fast because they use an atomic CPU instruction, rather than relying on external locks
)

// CountMetric is a count-based metric and implements the Metric interface
type CountMetric struct {
	value  *Count
	values map[string]map[string]*Count // map of [tagName][tagValue]*Count
}

func NewCountMetric() *CountMetric {
	return &CountMetric{}
}

func (c *CountMetric) AddSample(sample *MetricSample) error {
	if c.value == nil {
		c.value = NewCount()
	}
	if c.values == nil {
		c.values = make(map[string]map[string]*Count)
	}

	c.value.AddSample(sample)
	for _, tag := range sample.Tags {
		if _, ok := c.values[tag.Name]; !ok {
			c.values[tag.Name] = make(map[string]*Count)
		}
		if _, ok := c.values[tag.Name][tag.Value]; !ok {
			c.values[tag.Name][tag.Value] = NewCount()
		}
		c.values[tag.Name][tag.Value].AddSample(sample)
	}
	return nil
}

// Flush computes the metric and clears it. It returns error if the metric is not sampled
func (c *CountMetric) Flush(timestamp int64) (*ComputedMetric, error) {
	if c.value == nil {
		return nil, errors.NewUnsampledMetricError()
	}
	computedMetric := &ComputedMetric{
		Value:     c.value.Flush(),
		Timestamp: timestamp,
	}

	var tagGroups []*ComputedMetric
	for tagName, countPerTagValue := range c.values {
		var groupValue int64
		tagValueGroups := make([]*ComputedMetric, 0)
		for tagValue, count := range countPerTagValue {
			val := count.Flush()
			groupValue += val
			tagValueGroups = append(tagValueGroups, &ComputedMetric{
				Name:      tagValue,
				Value:     val,
				Timestamp: timestamp,
			})
		}
		tagGroups = append(tagGroups, &ComputedMetric{
			Name:      tagName,
			Value:     groupValue,
			Timestamp: timestamp,
			Groups:    tagValueGroups,
		})

	}
	computedMetric.Groups = tagGroups
	return computedMetric, nil
}

// Count is a thread-safe counter used to count the number of events that occur between 2 flushes
type Count struct {
	value int64
}

func NewCount() *Count {
	return &Count{}
}

func (c *Count) AddSample(sample *MetricSample) {
	atomic.AddInt64(&c.value, sample.Value)
}

func (c *Count) Flush() int64 {
	val := atomic.SwapInt64(&c.value, 0)
	return val
}
