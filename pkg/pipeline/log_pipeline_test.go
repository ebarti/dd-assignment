package pipeline

import (
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/ebarti/dd-assignment/pkg/errors"
	"github.com/ebarti/dd-assignment/pkg/logs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLogPipeline(t *testing.T) {
	defer goleak.VerifyNone(t)
	lines := []string{
		"\"10.0.0.2\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.4\",\"-\",\"apache\",1549573860,\"GET /api/user HTTP/1.0\",200,1234",
		"\"10.0.0.2\",\"-\",\"apache\",1549573860,\"GET /api/help HTTP/1.0\",200,1234",
		"\"10.0.0.5\",\"-\",\"apache\",1549573860,\"GET /api/help HTTP/1.0\",200,1234",
	}
	topLevelAttributeNames := []string{"http", "rfc931", "authuser", "request", "bytes"}
	httpAttributeNames := []string{"path", "method", "protocol"}
	pathAttributeNames := []string{"uri", "section", "subsection"}

	var msgs []*common.Message
	for _, line := range lines {
		msgs = append(msgs, &common.Message{
			Content:            []byte(line),
			Origin:             "test",
			IngestionTimestamp: time.Now().Unix(),
		})
	}

	inputChan := make(chan *common.Message)
	outputChan := make(chan *logs.ProcessedLog, len(msgs)+1)
	monitorChan := make(chan *logs.ProcessedLog, len(msgs)+1)
	logPipeline := NewLogPipeline(csvLogProcessor)
	logPipeline.OutputChan = outputChan
	logPipeline.From(inputChan)
	logPipeline.addMonitoredChannel(monitorChan)
	assert.NoError(t, logPipeline.Start())
	for _, msg := range msgs {
		inputChan <- msg
	}
	logPipeline.Stop()

	ii := 0

	// helper function to check the output and the monitored output
	outPutChecker := func(ii int, output *logs.ProcessedLog) {
		assert.Equal(t, "200", output.Status)
		assert.Equal(t, lines[ii], output.Message)
		assert.EqualValues(t, 1549573860, output.Timestamp)

		for _, topLevelAttribute := range topLevelAttributeNames {
			assert.Contains(t, output.Attributes, topLevelAttribute)
		}
		assert.EqualValues(t, "1234", output.Attributes["bytes"])

		httpAttributes := output.Attributes["http"].(map[string]interface{})
		for _, httpAttribute := range httpAttributeNames {
			assert.Contains(t, httpAttributes, httpAttribute)
		}
		assert.EqualValues(t, "GET", httpAttributes["method"])
		assert.EqualValues(t, "HTTP/1.0", httpAttributes["protocol"])

		pathAttributes := httpAttributes["path"].(map[string]interface{})
		for _, pathAttribute := range pathAttributeNames {
			assert.Contains(t, pathAttributes, pathAttribute)
		}
		if ii < 3 {
			assert.EqualValues(t, "/api/user", pathAttributes["uri"])
			assert.EqualValues(t, "api", pathAttributes["section"])
			assert.EqualValues(t, "user", pathAttributes["subsection"])
		} else {
			assert.EqualValues(t, "/api/help", pathAttributes["uri"])
			assert.EqualValues(t, "api", pathAttributes["section"])
			assert.EqualValues(t, "help", pathAttributes["subsection"])
		}
	}
	for output := range outputChan {
		// check that both output and monitored output are the same and match the expected output
		monitored := <-monitorChan
		outPutChecker(ii, output)
		outPutChecker(ii, monitored)
		ii++
	}
}

var csvLogProcessor LogProcessorFunc = func(msg *common.Message) (*logs.ProcessedLog, error) {
	header := "\"remotehost\",\"rfc931\",\"authuser\",\"date\",\"request\",\"status\",\"bytes\""
	headerLen := len(strings.Split(header, ","))
	content := string(msg.Content)
	if content == header {
		return nil, errors.NewInvalidLogLineError(content)
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
	return log, nil
}
