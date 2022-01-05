package backend

import (
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"log"
)

type Service struct {
	logPipeline      *pipeline.LogPipeline
	metricsPipeline  *metrics.MetricsPipeline
	metricAggregator *metrics.MetricAggregator
	monitors         []*monitors.LogMonitor
	logger           *log.Logger
}

func NewService(
	interval int64,
	logProcessor pipeline.LogProcessorFunc,
	customMetrics []*metrics.CustomMetricPipeline,
	monitorConfigs []*monitors.LogMonitorConfig,
	logger *log.Logger,
) *Service {
	aggregator := metrics.NewMetricAggregator(logger, interval)
	metricsPipeline := metrics.NewMetricsPipeline(customMetrics)
	logPipeline := pipeline.NewLogPipeline(logProcessor)
	aggregator.From(metricsPipeline.OutputChan)
	metricsPipeline.From(logPipeline.OutputChan)
	var m []*monitors.LogMonitor
	if len(monitorConfigs) > 0 {
		for _, config := range monitorConfigs {
			m = append(m, monitors.NewLogMonitor(config, logger))
		}
		logPipeline.AddMonitors(m)
	}

	return &Service{
		logPipeline:      logPipeline,
		metricsPipeline:  metricsPipeline,
		metricAggregator: aggregator,
		logger:           logger,
		monitors:         m,
	}
}

func (s *Service) From(input chan *common.Message) {
	s.logPipeline.From(input)
}

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
	return nil
}

func (s *Service) Stop() {
	// stop propagates down the pipeline
	s.logPipeline.Stop()
}
