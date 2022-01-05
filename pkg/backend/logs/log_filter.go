package logs

import (
	"github.com/ebarti/dd-assignment/pkg/backend/errors"
	"log"
	"strings"
)

type LogFilter struct {
	matchAll   bool
	status     *string
	host       *string
	service    *string
	attributes map[string]string
}

// NewLogFilter parses the query string and returns an LogFilter
// filter is of the format "service:MyService status:200 @http.path.section:mysection"
// NOTE:
// 	 Filtering for timestamp is NOT supported
// 	 Filtering for message content is NOT supported - e.g. querying datadog logs with "myQuery"
//   Wildcard filtering is NOT supported
func NewLogFilter(query string) *LogFilter {
	a := &LogFilter{}
	if err := a.build(query); err != nil {
		log.Fatalf("could not build aggregation filter query: %v", err)
	}
	return a
}

func (a *LogFilter) build(query string) error {
	if query == "*" {
		a.matchAll = true
		return nil
	}
	querySplit := strings.Split(query, " ")
	if len(querySplit) == 0 {
		return errors.NewInvalidAggregationQueryError(query)
	}
	for _, q := range querySplit {
		if strings.HasPrefix(q, "status:") {
			trimmedQuery := strings.TrimSuffix(q, "status:")
			a.status = &trimmedQuery
		} else if strings.HasPrefix(q, "host:") {
			trimmedQuery := strings.TrimSuffix(q, "host:")
			a.host = &trimmedQuery
		} else if strings.HasPrefix(q, "service:") {
			trimmedQuery := strings.TrimSuffix(q, "service:")
			a.service = &trimmedQuery
		} else if strings.HasPrefix(q, "@") {
			if !strings.Contains(q, ":") {
				return errors.NewInvalidAggregationQueryError(query)
			}
			if a.attributes == nil {
				a.attributes = make(map[string]string)
			}
			attributeSplit := strings.Split(q, ":")
			if len(attributeSplit) != 2 {
				return errors.NewInvalidAggregationQueryError(query)
			}
			attributePath := strings.TrimPrefix(attributeSplit[0], "@")
			attributeValue := attributeSplit[1]
			a.attributes[attributePath] = attributeValue
		} else {
			return errors.NewInvalidAggregationQueryError(query)
		}
	}
	return nil
}

func (a *LogFilter) Matches(log *ProcessedLog) bool {
	if a.matchAll {
		return true
	}
	return a.matchTopLevel(log) && a.matchAttributes(log)
}

func (a *LogFilter) matchTopLevel(log *ProcessedLog) bool {
	if a.status != nil && *a.status != log.Status {
		return false
	}
	if a.host != nil && *a.host != log.Host {
		return false
	}
	if a.service != nil && *a.service != log.Service {
		return false
	}
	return true
}

func (a *LogFilter) matchAttributes(log *ProcessedLog) bool {
	for attrPath, attrValue := range a.attributes {
		if !log.HasAttributeWithValue(attrPath, attrValue) {
			return false
		}
	}
	return true
}
