package pkg

import (
	"github.com/ebarti/dd-assignment/pkg/metrics"
	"github.com/ebarti/dd-assignment/pkg/monitors"
	"github.com/ebarti/dd-assignment/pkg/pipeline"
	"github.com/ebarti/dd-assignment/pkg/reader"
	"log"
	"os"
	"os/signal"
)

// Service is the main struct of the service that holds all the components created to solve this Datadog's take home exercise
type Service struct {
	reader           *reader.FileReader
	logPipeline      *pipeline.LogPipeline
	metricsPipeline  *metrics.MetricsPipeline
	metricAggregator *metrics.MetricAggregator
	monitors         []*monitors.LogMonitor
	sigChan          chan os.Signal
}

// NewService creates a new Service
func NewService(
	filePath string,
	interval int64,
	logProcessor pipeline.LogProcessorFunc,
	customMetrics []*metrics.CustomMetricPipeline,
	monitorConfigs []*monitors.LogMonitorConfig,
	logger *log.Logger,
) *Service {
	r := reader.NewFileReader(filePath, logger)
	logPipeline := pipeline.NewLogPipeline(logProcessor)
	logPipeline.From(r.OutputChan)
	metricsPipeline := metrics.NewMetricsPipeline(customMetrics)
	metricsPipeline.From(logPipeline.OutputChan)
	aggregator := metrics.NewMetricAggregator(logger, interval)
	aggregator.From(metricsPipeline.OutputChan)
	var m []*monitors.LogMonitor
	if len(monitorConfigs) > 0 {
		for _, config := range monitorConfigs {
			m = append(m, monitors.NewLogMonitor(config, logger))
		}
		logPipeline.AddMonitors(m)
	}

	return &Service{
		reader:           r,
		logPipeline:      logPipeline,
		metricsPipeline:  metricsPipeline,
		metricAggregator: aggregator,
		monitors:         m,
		sigChan:          make(chan os.Signal, 1),
	}
}

// CancelOnSignal : set the signal to be used to cancel the Service
func (s *Service) CancelOnSignal(signals ...os.Signal) *Service {
	if len(signals) > 0 {
		s.waitForSignal(signals...)
	}
	return s
}

// Start : start the service
func (s *Service) Start() error {
	// start services backwards
	if err := s.metricAggregator.Start(); err != nil {
		return err
	}
	if err := s.metricsPipeline.Start(); err != nil {
		return err
	}
	for _, m := range s.monitors {
		if err := m.Start(); err != nil {
			return err
		}
	}
	if err := s.logPipeline.Start(); err != nil {
		return err
	}
	if err := s.reader.Start(); err != nil {
		return err
	}
	return nil
}

// Wait : wait for the service to be cancelled
func (s *Service) Wait() {
	for !s.IsStopped() {
		// block
	}
	close(s.sigChan)
}

// IsStopped : check if the service is stopped
func (s *Service) IsStopped() bool {
	allStopped := s.reader.IsStopped() && s.logPipeline.IsStopped() && s.metricsPipeline.IsStopped() && s.metricAggregator.IsStopped()
	if allStopped {
		for _, m := range s.monitors {
			if !m.IsStopped() {
				allStopped = false
				break
			}
		}
	}
	return allStopped
}

// Stop : stop the service
func (s *Service) Stop() {
	// stop propagates down the pipeline
	s.reader.Stop()
	close(s.sigChan)
}

// waitForSignal : wait for a os.Signal to be received
func (s *Service) waitForSignal(signals ...os.Signal) {
	go func() {
		signal.Notify(s.sigChan, signals...)
		<-s.sigChan
		s.reader.Stop()
	}()
}
