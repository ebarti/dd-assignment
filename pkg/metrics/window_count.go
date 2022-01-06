package metrics

import (
	"github.com/ebarti/dd-assignment/pkg/errors"
	"github.com/gammazero/deque"
	"sync"
)

// WindowCountMetric is a metric that keeps track of the number of items in a sliding window.
type WindowCountMetric struct {
	timeWindow int64
	value      *WindowCount
	values     map[string]map[string]*WindowCount // map of [tagName][tagValue]*Count
}

// NewWindowCountMetric creates a new WindowCountMetric with the given time window.
func NewWindowCountMetric(timeWindow int64) *WindowCountMetric {
	return &WindowCountMetric{
		timeWindow: timeWindow,
	}
}

// AddSample adds a sample to the metric.
func (c *WindowCountMetric) AddSample(sample *MetricSample) error {
	if c.value == nil {
		c.value = NewWindowCount(c.timeWindow)
	}
	if c.values == nil {
		c.values = make(map[string]map[string]*WindowCount)
	}

	c.value.AddSample(sample)
	for _, tag := range sample.Tags {
		if _, ok := c.values[tag.Name]; !ok {
			c.values[tag.Name] = make(map[string]*WindowCount)
		}
		if _, ok := c.values[tag.Name][tag.Value]; !ok {
			c.values[tag.Name][tag.Value] = NewWindowCount(c.timeWindow)
		}
		c.values[tag.Name][tag.Value].AddSample(sample)
	}
	return nil
}

// Flush computes the metric and clears it. It returns error if the metric is not sampled
func (c *WindowCountMetric) Flush(timestamp int64) (*ComputedMetric, error) {
	if c.value == nil {
		return nil, errors.NewUnsampledMetricError()
	}
	computedMetric := &ComputedMetric{
		Value:     c.value.Flush(timestamp),
		Timestamp: timestamp,
	}

	var tagGroups []*ComputedMetric
	for tagName, countPerTagValue := range c.values {
		var groupValue int64
		tagValueGroups := make([]*ComputedMetric, 0)
		for tagValue, count := range countPerTagValue {
			val := count.Flush(timestamp)
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

// WindowCount is a thread-safe windowed count.
type WindowCount struct {
	Samples     *deque.Deque
	value       int64
	WindowWidth int64
	mu          sync.Mutex
}

// NewWindowCount creates a new WindowCount.
func NewWindowCount(windowWidth int64) *WindowCount {
	return &WindowCount{
		WindowWidth: windowWidth,
		Samples:     deque.New(),
	}
}

// AddSample adds a sample to the windowed counter.
func (w *WindowCount) AddSample(sample *MetricSample) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Samples.PushBack(sample)
	w.value += sample.Value
}

// Flush flushes the windowed counter. It will clear the entries older than the specified timestamp
// minus the window width, and return the re-computed value
func (w *WindowCount) Flush(timestamp int64) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	cutoff := timestamp - w.WindowWidth // cutoff included in the window
	for {
		if w.Samples.Len() == 0 || w.Samples.Front().(*MetricSample) == nil || w.Samples.Front().(*MetricSample).Timestamp >= cutoff {
			break
		}
		if w.Samples.Front().(*MetricSample) != nil && w.Samples.Front().(*MetricSample).Timestamp < cutoff {
			w.value -= w.Samples.PopFront().(*MetricSample).Value
		}
	}
	return w.value
}
