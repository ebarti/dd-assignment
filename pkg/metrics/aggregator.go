package metrics

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// MetricAggregator is the engine that collects metrics and reports them out every interval.
type MetricAggregator struct {
	logger            *log.Logger
	InputChan         chan []*MetricSample // all metrics in the array will have the same timestamp as they come from the same log
	interval          int64
	firstSampled      int64
	lastFlushed       int64
	currentTime       int64
	MetricsByInterval map[int64]map[string]Metric // map of [interval][metricName]Metric
	done              chan struct{}
	isDone            uint32
}

// NewMetricAggregator creates a new MetricAggregator.
func NewMetricAggregator(logger *log.Logger, interval int64) *MetricAggregator {
	return &MetricAggregator{
		logger:            logger,
		InputChan:         make(chan []*MetricSample),
		interval:          interval,
		MetricsByInterval: make(map[int64]map[string]Metric),
		done:              make(chan struct{}),
	}
}

// From is used to set the input channel for the MetricAggregator.
func (s *MetricAggregator) From(inputChan chan []*MetricSample) {
	s.InputChan = inputChan
}

// Start starts the MetricAggregator.
func (s *MetricAggregator) Start() error {
	go s.run()
	return nil
}

// Stop stops the MetricAggregator.
func (s *MetricAggregator) Stop() {
	if atomic.CompareAndSwapUint32(&s.isDone, 0, 1) {
		close(s.InputChan)
		<-s.done
	}
}

// IsStopped is used to verify that the MetricAggregator has stopped.
func (s *MetricAggregator) IsStopped() bool {
	return atomic.LoadUint32(&s.isDone) == 1
}

// cleanUp is used to set the stopped state if the input channel was closed.
func (s *MetricAggregator) cleanUp() {
	atomic.StoreUint32(&s.isDone, 1)
	close(s.done)
}

// run is the main loop of the MetricAggregator.
func (s *MetricAggregator) run() {
	defer s.cleanUp()
	for input := range s.InputChan {
		if len(input) == 0 {
			continue
		}
		wg := sync.WaitGroup{}
		wg.Add(2)
		// spawn a goroutine to update the samples
		go func() {
			defer wg.Done()
			s.addSamples(input)
		}()
		// spawn a goroutine to update the current time
		go func() {
			defer wg.Done()
			// all samples have the same timestamp as they come from the same log -> just get the first
			s.flushStats(input[0].Timestamp)
		}()
		wg.Wait()
	}
}

// addSamples is a helper function that adds the samples to the MetricAggregator.
func (s *MetricAggregator) addSamples(samples []*MetricSample) {
	for _, sample := range samples {
		bucket := s.getBucket(sample.Timestamp)
		if _, ok := s.MetricsByInterval[bucket]; !ok {
			s.MetricsByInterval[bucket] = make(map[string]Metric)
		}
		if _, ok := s.MetricsByInterval[bucket][sample.Name]; !ok {
			s.MetricsByInterval[bucket][sample.Name] = NewCountMetric()
		}
		s.MetricsByInterval[bucket][sample.Name].AddSample(sample)
	}
}

// flushStats computes the statistics for a given interval if they have not been flushed yet.
func (s *MetricAggregator) flushStats(timestamp int64) {
	intervalStart := atomic.LoadInt64(&s.lastFlushed)
	if timestamp < intervalStart+s.interval {
		return
	}
	intervalEnd := atomic.AddInt64(&s.lastFlushed, s.interval)
	metricsByName := s.MetricsByInterval[s.getBucket(timestamp)-1]
	if len(metricsByName) == 0 {
		return
	}
	var computedMetrics []*ComputedMetric
	for name, metric := range metricsByName {
		computedMetric, err := metric.Flush(timestamp)
		if err != nil {
			log.Printf("error while processing metric %s: %s", name, err)
			continue
		}
		computedMetric.Name = name
		// compute statistics for each series
		computedMetrics = append(computedMetrics, computedMetric)
	}
	// flush stats to the writer
	s.logger.Println(s.format(intervalStart, intervalEnd, timestamp, computedMetrics))
}

// format is a helper function that formats the computed metrics into a string.
func (s *MetricAggregator) format(start, end, timestamp int64, computedMetrics []*ComputedMetric) string {
	var buf bytes.Buffer
	flushTime := time.Unix(timestamp, 0)
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)

	fmt.Fprintf(&buf, "[%v] Statistics for time interval %v-%v\n", flushTime, startTime, endTime)

	for _, metric := range computedMetrics {
		metric.Render(&buf)
		buf.WriteRune('\n')
	}
	return buf.String()
}

// getBucket gets the interval bucket for a given timestamp.
func (s *MetricAggregator) getBucket(timestamp int64) int64 {
	// no need to atomically read the firstSampled or the interval as they are only written once
	if s.firstSampled == 0 {
		s.firstSampled = timestamp
		// init last flushed time
		atomic.StoreInt64(&s.lastFlushed, timestamp)
	}
	return (timestamp - s.firstSampled) / s.interval
}
