package backend

import (
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/backend/errors"
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/metrics"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"strconv"
	"strings"
	"time"
)

func GetCsvLogMonitorConfig(interval, threshold int64) *monitors.LogMonitorConfig {
	return &monitors.LogMonitorConfig{
		Name:           "High traffic monitor",
		TimeWindow:     interval,
		Filter:         "*",
		AlertThreshold: threshold,
		AlertTemplate:  "High traffic generated an alert - hits {{value}}, triggered at {{time}}",
		AlertTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
			return map[string]string{
				"value": strconv.FormatInt(m.Value, 10),
				"time":  fmt.Sprintf("%v", time.Unix(m.Timestamp, 0)),
			}
		},
		RecoveryTemplate:  "Recovered from high traffic at time {{time}}",
		RecoveryThreshold: interval,
		RecoveryTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
			return map[string]string{
				"time": fmt.Sprintf("%v", time.Unix(m.Timestamp, 0)),
			}
		},
	}
}

func GetCsvLogProcessingFunc() pipeline.LogProcessorFunc {
	return func(msg *common.Message) (*logs.ProcessedLog, error) {
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
}

func GetCsvCustomMetricsPipelines(interval int64) *metrics.CustomMetricPipeline {
	return metrics.NewCustomMetricPipeline("DD exercise", "*", interval, nil, []string{"http.path.section", "http.path.subsection", "status"})
}
