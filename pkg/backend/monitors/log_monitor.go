package monitors

import (
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/hoisie/mustache"
	"io"
	"log"
	"sync/atomic"
)

type ContextExtractorFunc func(*metrics.ComputedMetric) map[string]string

type LogMonitorConfig struct {
	name                        string
	timeWindow                  int64
	filter                      string
	alertThreshold              int64
	alertTemplate               string
	alertTemplateContextFunc    ContextExtractorFunc
	recoveryThreshold           int64
	recoveryTemplate            string
	recoveryTemplateContextFunc ContextExtractorFunc
}

// only counter monitors are supported - in Datadog's terms, only .rollup("count") is supported
// no group by supported here
type LogMonitor struct {
	name                        string
	timeWindow                  int64
	alertThreshold              int64
	alertTemplate               *mustache.Template // High traffic generated an alert - hits {{value}}, triggered at {{time}}
	alertTemplateContextFunc    ContextExtractorFunc
	recoveryThreshold           int64
	recoveryTemplate            *mustache.Template
	recoveryTemplateContextFunc ContextExtractorFunc
	customMetric                *metrics.CustomMetricPipeline
	metric                      *metrics.WindowCountMetric
	IsInAlert                   bool
	lastChecked                 int64
	writer                      io.Writer
	InputChan                   chan *logs.ProcessedLog
	done                        chan struct{}
	isDone                      uint32
}

func NewLogMonitor(config *LogMonitorConfig, writer io.Writer) *LogMonitor {
	aTmpl, err := mustache.ParseString(config.alertTemplate)
	if err != nil {
		// we should not panic here, but we will for this exercise
		log.Fatalf("Failed to parse alert template: %s", err)
	}
	recoveryTemplate := config.recoveryTemplate
	if config.recoveryTemplate == "" {
		recoveryTemplate = "[RECOVERED] " + config.alertTemplate
	}
	rTmpl, err := mustache.ParseString(recoveryTemplate)
	if err != nil {
		// we should not panic here, but we will for this exercise
		log.Fatalf("Failed to parse recovery template: %s", err)
	}
	rFunc := config.recoveryTemplateContextFunc
	if rFunc == nil {
		rFunc = config.alertTemplateContextFunc
	}
	return &LogMonitor{
		name:                        config.name,
		writer:                      writer,
		alertThreshold:              config.alertThreshold,
		alertTemplate:               aTmpl,
		alertTemplateContextFunc:    config.alertTemplateContextFunc,
		recoveryThreshold:           config.recoveryThreshold,
		recoveryTemplate:            rTmpl,
		recoveryTemplateContextFunc: rFunc,
		customMetric:                metrics.NewCustomMetricPipeline(config.name, config.filter, config.timeWindow, nil, nil),
		metric:                      metrics.NewWindowCountMetric(config.timeWindow),
		timeWindow:                  config.timeWindow,
		InputChan:                   make(chan *logs.ProcessedLog, 100),
		done:                        make(chan struct{}),
	}
}

func (m *LogMonitor) Start() error {
	go m.monitor()
	return nil
}

func (m *LogMonitor) Stop() {
	if atomic.CompareAndSwapUint32(&m.isDone, 0, 1) {
		close(m.InputChan)
		<-m.done
	}
}

func (m *LogMonitor) monitor() {
	defer m.cleanUp()
	for input := range m.InputChan {
		timestamp := input.Timestamp
		shouldFlush := timestamp > m.lastChecked
		metric := m.customMetric.Compute(input)
		if metric != nil {
			m.metric.AddSample(metric)
			// as samples might come out of order, always flush if this sample matches the alert
			shouldFlush = true
		}
		if !shouldFlush {
			continue
		}
		computedMetric, err := m.metric.Flush(timestamp)
		if err != nil {
			continue
		}
		val, err := computedMetric.GetValue(nil) // no tags supported in monitoring yet
		if err != nil {
			continue
		}

		if val > m.alertThreshold && !m.IsInAlert {
			m.setInAlert()
			m.writer.Write([]byte(m.alertTemplate.Render(m.alertTemplateContextFunc(computedMetric))))
		} else if val <= m.recoveryThreshold && m.IsInAlert {
			m.clearAlert()
			m.writer.Write([]byte(m.recoveryTemplate.Render(m.recoveryTemplateContextFunc(computedMetric))))
		}
	}
}

func (m *LogMonitor) cleanUp() {
	atomic.StoreUint32(&m.isDone, 1)
	close(m.done)
}

func (m *LogMonitor) setInAlert() {
	m.IsInAlert = true
}

func (m *LogMonitor) clearAlert() {
	m.IsInAlert = false
}

func (m *LogMonitor) isInAlert() bool {
	return m.IsInAlert == true
}
