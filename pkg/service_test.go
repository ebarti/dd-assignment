package pkg

import (
	"bytes"
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/ebarti/dd-assignment/pkg/errors"
	"github.com/ebarti/dd-assignment/pkg/logs"
	"github.com/ebarti/dd-assignment/pkg/metrics"
	"github.com/ebarti/dd-assignment/pkg/monitors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestService(t *testing.T) {
	defer goleak.VerifyNone(t)
	filePath := "../test_resources/service_test_csv.txt"
	//  strconv.FormatInt(, 10)
	expectedOutputLines := []string{
		"High traffic generated an alert - hits 3, triggered at 101",
		"Recovered from high traffic at time 105",
		"High traffic generated an alert - hits 3, triggered at 106",
		"Recovered from high traffic at time 109",
	}
	_ = expectedOutputLines
	buf := bytes.Buffer{}
	logger := log.New(&buf, "", 0)

	service := NewService(filePath, 2, logPipelineForTest, []*metrics.CustomMetricPipeline{customMetricPipelineForTest}, []*monitors.LogMonitorConfig{logMonitorConfigForTest}, logger).CancelOnSignal(os.Interrupt, os.Kill)
	assert.NoError(t, service.Start())
	service.Wait()
	output := buf.String()
	for _, line := range expectedOutputLines {
		// as the log monitoring runs on different threads as the statistics, we cannot guarantee the order of the output
		assert.Truef(t, strings.Contains(output, line), "expected output to contain %s", line)
	}
}

var logPipelineForTest = func(msg *common.Message) (*logs.ProcessedLog, error) {
	header := "\"remotehost\",\"rfc931\",\"authuser\",\"date\",\"request\",\"status\",\"bytes\""
	headerLen := len(strings.Split(header, ","))
	content := string(msg.Content)
	if content == header {
		return nil, nil
	}
	log := &logs.ProcessedLog{
		Host:       msg.Origin,
		Message:    content,
		Attributes: make(map[string]interface{}),
	}
	splitContent := strings.Split(content, ",")
	if len(splitContent) < headerLen {
		return nil, errors.NewInvalidCsvLogFormatError(len(splitContent), headerLen)
	}
	// trim quotes
	for ii := 0; ii < len(splitContent); ii++ {
		splitContent[ii] = strings.TrimPrefix(splitContent[ii], "\"")
		splitContent[ii] = strings.TrimSuffix(splitContent[ii], "\"")
	}
	// host
	log.Host = splitContent[0]
	// date
	date, err := strconv.ParseInt(splitContent[3], 10, 64)
	if err != nil {
		return nil, errors.NewUnableToParseDateError(splitContent[3], err)
	}
	log.Timestamp = date
	// attributes
	request := splitContent[4]
	log.Attributes["rfc931"] = splitContent[1]
	log.Attributes["authuser"] = splitContent[2]
	log.Attributes["request"] = request
	log.Status = splitContent[5]
	log.Attributes["bytes"] = splitContent[6]

	// parse request. Example: GET /api/user HTTP/1.0
	splitRequest := strings.Split(request, " ")
	if len(splitRequest) < 3 {
		return nil, errors.NewInvalidRequestFormatError(request)
	}
	httpAttributes := make(map[string]interface{})
	httpAttributes["method"] = splitRequest[0]
	httpAttributes["protocol"] = splitRequest[2]

	// Parse path attributes. Example: /api/user
	uri := splitRequest[1]
	splitPath := strings.Split(uri, "/")
	if len(splitPath) < 2 {
		return nil, errors.NewInvalidRequestFormatError(request)
	}
	pathAttributes := make(map[string]interface{})
	pathAttributes["uri"] = uri
	pathAttributes["section"] = splitPath[1]
	if len(splitPath) > 2 {
		pathAttributes["subsection"] = splitPath[2]
	}
	httpAttributes["path"] = pathAttributes
	log.Attributes["http"] = httpAttributes
	log.Message = string(msg.Content)
	return log, nil
}

var customMetricPipelineForTest = metrics.NewCustomMetricPipeline("DD exercise", "*", 2, nil, []string{"http.path.section", "http.path.subsection", "status"})

var logMonitorConfigForTest = &monitors.LogMonitorConfig{
	Name:           "High traffic monitor",
	TimeWindow:     2,
	Filter:         "*",
	AlertThreshold: 2,
	AlertTemplate:  "High traffic generated an alert - hits {{value}}, triggered at {{time}}",
	AlertTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
		return map[string]string{
			"value": strconv.FormatInt(m.Value, 10),
			"time":  fmt.Sprintf("%v", strconv.FormatInt(m.Timestamp, 10)),
		}
	},
	RecoveryTemplate:  "Recovered from high traffic at time {{time}}",
	RecoveryThreshold: 2,
	RecoveryTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
		return map[string]string{
			"time": fmt.Sprintf("%v", strconv.FormatInt(m.Timestamp, 10)),
		}
	},
}
