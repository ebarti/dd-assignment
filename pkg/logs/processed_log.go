package logs

import (
	"strconv"
	"strings"
)

type ProcessedLog struct {
	Timestamp  int64
	Status     string
	Host       string
	Service    string
	Message    string
	Attributes map[string]interface{}
}

func (l *ProcessedLog) HasAttributeWithValue(path string, want string) bool {
	has := l.GetAttribute(path)
	if has == nil {
		return false
	}
	return *has == want
}

func (l *ProcessedLog) GetAttribute(attribute string) *string {
	lowerCaseAttribute := strings.ToLower(attribute)
	if lowerCaseAttribute == "status" {
		return &l.Status
	} else if lowerCaseAttribute == "host" {
		return &l.Host
	} else if lowerCaseAttribute == "service" {
		return &l.Service
	} else if lowerCaseAttribute == "message" {
		return &l.Message
	} else if lowerCaseAttribute == "timestamp" {
		ts := strconv.Itoa(int(l.Timestamp))
		return &ts
	}
	if l.Attributes == nil {
		return nil
	}
	attrPathSplit := strings.Split(attribute, ".")
	return l.getAttributeAtPath(l.Attributes, attrPathSplit)
}

func (l *ProcessedLog) getAttributeAtPath(got interface{}, attrPathSplit []string) *string {
	v, ok := got.(map[string]interface{})
	if !ok { // Catch (this condition should never happen)
		if strVal, ok := got.(string); ok {
			return &strVal
		}
		return nil
	}

	if val, ok := v[attrPathSplit[0]]; ok {
		if len(attrPathSplit) == 1 {
			if strVal, ok := val.(string); ok {
				return &strVal
			}
			return nil
		}
		return l.getAttributeAtPath(val, attrPathSplit[1:])
	}
	return nil
}
