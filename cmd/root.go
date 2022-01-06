package cmd

import (
	"fmt"
	"github.com/ebarti/dd-assignment/pkg"
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/ebarti/dd-assignment/pkg/errors"
	"github.com/ebarti/dd-assignment/pkg/logs"
	"github.com/ebarti/dd-assignment/pkg/metrics"
	"github.com/ebarti/dd-assignment/pkg/monitors"
	"github.com/ebarti/dd-assignment/pkg/pipeline"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	FileFlag              = "file"
	StatPrintIntervalFlag = "interval"
	AlertThresholdFlag    = "threshold"
	AlertTimeWindow       = "window"
)

// Note: This file was bootstrapped using cobra init.

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "dd-assignment",
		Short: "Eloi Barti's implementation of Datadog's take home assignment",
		Long:  `Eloi Barti's implementation of Datadog's take home assignment in go.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: runRootCmd,
	}
	filePath          string
	statPrintInterval int64
	alertThreshold    int64
	alertTimeWindow   int64
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, FileFlag, "f", "", "Path to the CSV file to process")
	rootCmd.MarkFlagRequired(FileFlag)
	rootCmd.Flags().Int64VarP(&statPrintInterval, StatPrintIntervalFlag, "i", 10, "Interval at which to output statistics in seconds")
	rootCmd.Flags().Int64VarP(&alertThreshold, AlertThresholdFlag, "t", 10, "Number of requests per second that, once aggregated over the time window, will trigger an alert")
	rootCmd.Flags().Int64VarP(&alertTimeWindow, AlertTimeWindow, "w", 2*60, "Time window in seconds to aggregate requests over")
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	service := pkg.NewService(
		filePath,
		statPrintInterval,
		GetCsvLogProcessingFunc(),
		GetCsvCustomMetricsPipelines(statPrintInterval),
		GetCsvLogMonitorConfig(alertTimeWindow, alertThreshold),
		log.New(os.Stdout, "", 0),
	)
	if err := service.Start(); err != nil {
		return err
	}
	// Wait until the service is done
	service.Wait()
	return nil
}

// GetCsvLogProcessingFunc is the definition of the "high traffic monitor"
func GetCsvLogMonitorConfig(timeWindow, threshold int64) []*monitors.LogMonitorConfig {
	return []*monitors.LogMonitorConfig{
		{
			Name:           "High traffic monitor",
			TimeWindow:     timeWindow, // two minutes
			Filter:         "*",
			AlertThreshold: timeWindow * threshold,
			AlertTemplate:  "High traffic generated an alert - hits {{value}}, triggered at {{time}}",
			AlertTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
				return map[string]string{
					"value": strconv.FormatInt(m.Value, 10),
					"time":  fmt.Sprintf("%v", time.Unix(m.Timestamp, 0)),
				}
			},
			RecoveryTemplate:  "Recovered from high traffic at time {{time}}",
			RecoveryThreshold: timeWindow * threshold,
			RecoveryTemplateContextFunc: func(m *metrics.ComputedMetric) map[string]string {
				return map[string]string{
					"time": fmt.Sprintf("%v", time.Unix(m.Timestamp, 0)),
				}
			},
		},
	}
}

// GetCsvLogProcessingFunc returns the custom log processing function for this exercise
func GetCsvLogProcessingFunc() pipeline.LogProcessorFunc {
	return func(msg *common.Message) (*logs.ProcessedLog, error) {
		header := "\"remotehost\",\"rfc931\",\"authuser\",\"date\",\"request\",\"status\",\"bytes\""
		headerLen := len(strings.Split(header, ","))
		content := string(msg.Content)
		// avoid creating a message for the header
		if content == header {
			return nil, errors.NewInvalidLogLineError(content) // expected
		}
		l := &logs.ProcessedLog{
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
		l.Host = splitContent[0]
		// date
		date, err := strconv.ParseInt(splitContent[3], 10, 64)
		if err != nil {
			return nil, errors.NewUnableToParseDateError(splitContent[3], err)
		}
		l.Timestamp = date
		// attributes
		request := splitContent[4]
		l.Attributes["rfc931"] = splitContent[1]
		l.Attributes["authuser"] = splitContent[2]
		l.Attributes["request"] = request
		l.Status = splitContent[5]
		l.Attributes["bytes"] = splitContent[6]

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
		l.Attributes["http"] = httpAttributes
		l.Message = string(msg.Content)
		return l, nil
	}
}

// GetCsvCustomMetricsPipelines returns the custom metrics pipelines for this exercise
// The statistics computed will be hits per section and subsection, as well as a count of status codes.
func GetCsvCustomMetricsPipelines(interval int64) []*metrics.CustomMetricPipeline {
	return []*metrics.CustomMetricPipeline{
		metrics.NewCustomMetricPipeline(
			"DD exercise",
			"*",
			interval,
			nil,
			[]string{"http.path.section", "http.path.subsection", "status"},
		),
	}
}
