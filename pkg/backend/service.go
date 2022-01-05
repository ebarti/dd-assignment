package backend

import (
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"io"
)

type Service struct {
	inputChan        chan *common.Message
	logPipeline      *pipeline.LogPipeline
	metricsPipeline  *metrics.MetricsPipeline
	metricAggregator *metrics.MetricAggregator
	writer           io.Writer
}

func NewService(
	interval int64,
	logProcessor pipeline.LogProcessorFunc,
	customMetrics []*metrics.CustomMetricPipeline,
	monitorConfigs []*monitors.LogMonitorConfig,
	writer io.Writer,
) *Service {

	aggregator := metrics.NewMetricAggregator(writer, interval)
	metricsPipeline := metrics.NewMetricsPipeline(customMetrics)
	logPipeline := pipeline.NewLogPipeline(logProcessor)

	if len(monitorConfigs) > 0 {
		var logMonitors []*monitors.LogMonitor
		for _, config := range monitorConfigs {
			logMonitors = append(logMonitors, monitors.NewLogMonitor(config, writer))
		}
		logPipeline.AddMonitors(logMonitors)
	}

	return &Service{
		inputChan:        make(chan *common.Message, 100),
		logPipeline:      logPipeline,
		metricsPipeline:  metricsPipeline,
		metricAggregator: aggregator,
		writer:           writer,
	}
}

func (s *Service) From(input chan *common.Message) {
	s.inputChan = input
}

func (s *Service) setUp() {
	s.metricAggregator.From(s.metricsPipeline.OutputChan)
	s.metricsPipeline.From(s.logPipeline.OutputChan)
	s.logPipeline.From(s.inputChan)
}
