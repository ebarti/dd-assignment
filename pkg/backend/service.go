package backend

import (
	common2 "github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"io"
)

type Service struct {
	InputChan        chan *common.Message
	logPipeline      *pipeline.LogPipeline
	metricsPipeline  *metrics.MetricsPipeline
	metricAggregator *metrics.MetricAggregator
	logMonitors      []*monitors.LogMonitor
	writer           io.Writer
}

func NewService(
	interval int64,
	InputChan chan *common.Message,
	logProcessor pipeline.LogProcessorFunc,
	customMetrics []*metrics.CustomMetricPipeline,
	monitorConfigs []*monitors.LogMonitorConfig,
	writer io.Writer,
) *Service {
	tmp := make(chan []*metrics.MetricSample)
	aggregator := metrics.NewMetricAggregator(writer, tmp, interval)

	out := make(chan *common2.ProcessedLog)
	logPipeline := pipeline.NewLogPipeline(InputChan, out, logProcessor)
	var logMonitors []*monitors.LogMonitor
	for _, config := range monitorConfigs {
		logMonitors = append(logMonitors, monitors.NewLogMonitor(config, writer))
	}
	metricsPipeline := metrics.NewMetricsPipeline(logPipeline.OutputChan, aggregator.InputChan, customMetrics)

	return &Service{
		InputChan:        make(chan *common.Message, 100),
		logPipeline:      logPipeline,
		metricsPipeline:  metricsPipeline,
		metricAggregator: aggregator,
		writer:           writer,
	}
}
