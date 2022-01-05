package metrics

import (
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
)

type MetricsPipeline struct {
	inputChan     chan *logs.ProcessedLog
	outputChan    chan []*MetricSample
	customMetrics []*CustomMetricPipeline
}

func NewMetricsPipeline(
	inputChan chan *logs.ProcessedLog,
	outputChan chan []*MetricSample,
	customMetrics []*CustomMetricPipeline,
) *MetricsPipeline {
	return &MetricsPipeline{
		inputChan:     inputChan,
		outputChan:    outputChan,
		customMetrics: customMetrics,
	}
}

func (m *MetricsPipeline) Start() error {
	go m.run()
	return nil
}

func (m *MetricsPipeline) Stop() {
	close(m.inputChan)
}

func (m *MetricsPipeline) run() {
	for input := range m.inputChan {
		var metrics []*MetricSample
		for _, metric := range m.customMetrics {
			metrics = append(metrics, metric.Compute(input))
		}
		m.outputChan <- metrics
	}
	close(m.outputChan)
}

type CustomMetricPipeline struct {
	name       string
	filter     *logs.LogFilter
	measure    *logs.LogMeasure
	groupBy    []string
	timeWindow int64
}

func NewCustomMetricPipeline(metricName string, filter string, timeWindow int64, measure *string, groupBy []string) *CustomMetricPipeline {
	return &CustomMetricPipeline{
		name:       metricName,
		filter:     logs.NewLogFilter(filter),
		measure:    logs.NewLogMeasure(measure),
		groupBy:    groupBy,
		timeWindow: timeWindow,
	}
}

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
