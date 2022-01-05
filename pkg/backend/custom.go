package backend

import (
	"github.com/ebarti/dd-assignment/pkg/backend/errors"
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/pipeline"
	"github.com/ebarti/dd-assignment/pkg/common"
	"strconv"
	"strings"
)

const (
	header = "\"remotehost\",\"rfc931\",\"authuser\",\"date\",\"request\",\"status\",\"bytes\""
)

var headerLen = len(strings.Split(header, ","))

var CsvLogProcessor pipeline.LogProcessorFunc = func(msg *common.Message) (*logs.ProcessedLog, error) {
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
	log.Attributes["status"] = splitContent[5]
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
