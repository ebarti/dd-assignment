package metrics

import (
	"github.com/ebarti/dd-assignment/pkg/logs"
	"sync/atomic"
)

// MetricsPipeline is a pipeline that computes metrics based on logs.ProcessedLog
type MetricsPipeline struct {
	inputChan     chan *logs.ProcessedLog
	OutputChan    chan []*MetricSample
	customMetrics []*CustomMetricPipeline
	done          chan struct{}
	isDone        uint32
}

// NewMetricsPipeline creates a new MetricsPipeline
func NewMetricsPipeline(customMetrics []*CustomMetricPipeline) *MetricsPipeline {
	return &MetricsPipeline{
		inputChan:     make(chan *logs.ProcessedLog),
		OutputChan:    make(chan []*MetricSample),
		customMetrics: customMetrics,
		done:          make(chan struct{}),
	}
}

// From is used to set the input channel for the MetricsPipeline
func (m *MetricsPipeline) From(inputChan chan *logs.ProcessedLog) {
	m.inputChan = inputChan
}

// Start starts the MetricsPipeline
func (m *MetricsPipeline) Start() error {
	go m.run()
	return nil
}

// Stop stops the MetricsPipeline
func (m *MetricsPipeline) Stop() {
	if atomic.CompareAndSwapUint32(&m.isDone, 0, 1) {
		close(m.inputChan)
		<-m.done
	}
}

// IsStopped is used to verify if the MetricsPipeline is stopped
func (m *MetricsPipeline) IsStopped() bool {
	return atomic.LoadUint32(&m.isDone) == 1
}

// run is the main loop of the MetricsPipeline
func (m *MetricsPipeline) run() {
	defer m.cleanUp()
	for input := range m.inputChan {
		var metrics []*MetricSample
		for _, metric := range m.customMetrics {
			metrics = append(metrics, metric.Compute(input))
		}
		m.OutputChan <- metrics
	}
}

// cleanUp closes the output channel and sets its stopped state
func (m *MetricsPipeline) cleanUp() {
	close(m.OutputChan)
	atomic.StoreUint32(&m.isDone, 1)
	close(m.done)
}

// CustomMetricPipeline is a customizable pipeline that computes metrics based on logs.ProcessedLog
type CustomMetricPipeline struct {
	name       string
	filter     *logs.LogFilter
	measure    *logs.LogMeasure
	groupBy    []string
	timeWindow int64
}

// NewCustomMetricPipeline creates a new CustomMetricPipeline with the given name, filter, measure, groupBy and timeWindow
func NewCustomMetricPipeline(metricName string, filter string, timeWindow int64, measure *string, groupBy []string) *CustomMetricPipeline {
	return &CustomMetricPipeline{
		name:       metricName,
		filter:     logs.NewLogFilter(filter),
		measure:    logs.NewLogMeasure(measure),
		groupBy:    groupBy,
		timeWindow: timeWindow,
	}
}

// Compute computes the metrics based on the given logs.ProcessedLog
func (s *CustomMetricPipeline) Compute(log *logs.ProcessedLog) *MetricSample {
	// Check if we need to process the log for this aggregate
	if !s.filter.Matches(log) {
		return nil
	}
	sample := &MetricSample{
		Name:      s.name,
		Timestamp: log.Timestamp,
	}
	// If we need to measure the log, do it
	var value int64
	value = 1
	if s.measure != nil {
		measuredLog := s.measure.Measure(log)
		if measuredLog != nil {
			value = *measuredLog
		} else {
			// Could not measure log
			return nil
		}
	}
	sample.Value = value
	// Enrich log with the required tags if we are grouping by
	if len(s.groupBy) > 0 {
		tags := []*Tag{}
		for _, g := range s.groupBy {
			val := log.GetAttribute(g)
			if val == nil {
				continue
			}
			tags = append(tags, &Tag{
				Name:  g,
				Value: *val,
			})
		}
		sample.Tags = tags
	}
	return sample
}
