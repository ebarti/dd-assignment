package metrics

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type MetricAggregator struct {
	writer            io.Writer
	InputChan         chan []*MetricSample // all metrics in the array will have the same timestamp as they come from the same log
	interval          int64
	firstSampled      int64
	lastFlushed       int64
	currentTime       int64
	MetricsByInterval map[int64]map[string]Metric // map of [interval][metricName]Metric
	done              chan struct{}
	isDone            uint32
}

func NewMetricAggregator(writer io.Writer, interval int64) *MetricAggregator {
	return &MetricAggregator{
		writer:            writer,
		InputChan:         make(chan []*MetricSample),
		interval:          interval,
		MetricsByInterval: make(map[int64]map[string]Metric),
		done:              make(chan struct{}),
	}
}

func (s *MetricAggregator) From(inputChan chan []*MetricSample) {
	s.InputChan = inputChan
}

func (s *MetricAggregator) Start() error {
	go s.run()
	return nil
}

func (s *MetricAggregator) Stop() {
	close(s.InputChan)
	if atomic.CompareAndSwapUint32(&s.isDone, 0, 1) {
		<-s.done
	}
}

func (s *MetricAggregator) run() {
	defer close(s.done)
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
	s.writer.Write(s.format(intervalStart, intervalEnd, timestamp, computedMetrics))
}

func (s *MetricAggregator) format(start, end, timestamp int64, computedMetrics []*ComputedMetric) []byte {
	var buf bytes.Buffer
	flushTime := time.Unix(timestamp, 0)
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)

	fmt.Fprintf(&buf, "[%v] Statistics for time interval %v-%v\n", flushTime, startTime, endTime)

	for _, metric := range computedMetrics {
		metric.Render(&buf)
		buf.WriteRune('\n')
	}
	buf.WriteRune('\n')
	return buf.Bytes()
}

func (s *MetricAggregator) getBucket(timestamp int64) int64 {
	// no need to atomically read the firstSampled or the interval as they are only written once
	if s.firstSampled == 0 {
		s.firstSampled = timestamp
		// init last flushed time
		atomic.StoreInt64(&s.lastFlushed, timestamp)
	}
	return (timestamp - s.firstSampled) / s.interval
}
