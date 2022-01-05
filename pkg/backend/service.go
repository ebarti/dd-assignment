package backend

import (
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"io"
)

type Service struct {
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
	aggregator.From(metricsPipeline.OutputChan)
	metricsPipeline.From(logPipeline.OutputChan)
	if len(monitorConfigs) > 0 {
		var logMonitors []*monitors.LogMonitor
		for _, config := range monitorConfigs {
			logMonitors = append(logMonitors, monitors.NewLogMonitor(config, writer))
		}
		logPipeline.AddMonitors(logMonitors)
	}

	return &Service{
		logPipeline:      logPipeline,
		metricsPipeline:  metricsPipeline,
		metricAggregator: aggregator,
		writer:           writer,
	}
}

func (s *Service) From(input chan *common.Message) {
	s.logPipeline.From(input)
}
