package monitors

import (
	"bytes"
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/logs"
	"github.com/ebarti/dd-assignment/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"log"
	"strconv"
	"strings"
	"testing"
)

func TestLogMonitor(t *testing.T) {
	defer goleak.VerifyNone(t)
	buf := bytes.Buffer{}
	logger := log.New(&buf, "", 0)
	inputChan := make(chan *logs.ProcessedLog)
	expectedOutputLines := []string{
		"High traffic generated an alert - hits 3, triggered at 101",
		"Recovered from high traffic at time 105",
		"High traffic generated an alert - hits 3, triggered at 106",
		"Recovered from high traffic at time 109",
	}
	logMonitorConfig := &LogMonitorConfig{
		Name:           "High traffic monitor",
		TimeWindow:     2,
		Filter:         "*",
		AlertThreshold: 2,
		AlertTemplate:  "High traffic generated an alert - hits {{value}}, triggered at {{time}}",
		AlertTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
			return map[string]string{
				"value": strconv.FormatInt(m.Value, 10),
				"time":  strconv.FormatInt(m.Timestamp, 10),
			}
		},
		RecoveryTemplate:  "Recovered from high traffic at time {{time}}",
		RecoveryThreshold: 2,
		RecoveryTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
			return map[string]string{
				"time": strconv.FormatInt(m.Timestamp, 10),
			}
		},
	}

	logMonitor := NewLogMonitor(logMonitorConfig, logger)
	logMonitor.InputChan = inputChan
	assert.NoError(t, logMonitor.Start())
	for _, l := range logsForMonitorTest {
		inputChan <- l
	}
	logMonitor.Stop()
	got := buf.String()
	fmt.Println(got)
	splitGot := strings.Split(got, "\n")
	// we append an extra \n everytime we render
	splitGot = splitGot[:len(splitGot)-1]
	for i, line := range splitGot {
		assert.Equal(t, expectedOutputLines[i], line)
	}
}

var logsForMonitorTest = []*logs.ProcessedLog{
	{
		Timestamp: 100,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 101,
		Status:    "201",
		Host:      "bHost",
	},
	{
		Timestamp: 101, // trigger alert
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 102,
		Status:    "301",
		Host:      "cHost",
	},
	{
		Timestamp: 103,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 105, // recover
		Status:    "401",
		Host:      "bHost",
	},
	{
		Timestamp: 106,
		Status:    "200",
		Host:      "aHost",
	},
	{
		Timestamp: 106, // retrigger
		Status:    "404",
		Host:      "aHost",
	},
	{
		Timestamp: 108,
		Status:    "204",
		Host:      "aHost",
	},
	{
		Timestamp: 109, // recover
		Status:    "204",
		Host:      "aHost",
	},
}
