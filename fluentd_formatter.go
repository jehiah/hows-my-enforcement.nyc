package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// severityMap maps values to be recognized by the Google Cloud Platform.
// https://cloud.google.com/logging/docs/agent/configuration#special-fields
// from google.golang.org/genproto/googleapis/logging/type
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
var severityMap = map[string]string{
	"panic":   "CRITICAL",
	"error":   "ERROR",
	"fatal":   "ALERT",
	"warning": "WARNING",
	"info":    "INFO",
	"debug":   "DEBUG",
	"trace":   "DEBUG",
}

// fluentdFormatter is similar to logrus.JSONFormatter w/ additions from
// https://github.com/joonix/log
type fluentdFormatter struct {
}

// https://cloud.google.com/logging/docs/agent/configuration#special-fields
type timestamp struct {
	// Represents seconds of UTC time since Unix epoch
	// 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to
	// 9999-12-31T23:59:59Z inclusive.
	Seconds int64 `json:"seconds,omitempty"`
	// Non-negative fractions of a second at nanosecond resolution. Negative
	// second values with fractions must still have non-negative nanos values
	// that count forward in time. Must be from 0 to 999,999,999
	// inclusive.
	Nanos int32 `json:"nanos,omitempty"`
	// contains filtered or unexported fields
}

// Format renders a single log entry
func (f *fluentdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	prefixFieldClashes(data, entry.HasCaller())

	data["timestamp"] = timestamp{
		Seconds: entry.Time.Unix(),
		Nanos:   int32(entry.Time.Nanosecond()),
	}
	data["severity"] = severityMap[entry.Level.String()]

	data["message"] = entry.Message
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		if funcVal != "" {
			data[logrus.FieldKeyFunc] = funcVal
		}
		if fileVal != "" {
			data[logrus.FieldKeyFile] = fileVal
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return b.Bytes(), nil
}

func prefixFieldClashes(data logrus.Fields, reportCaller bool) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
		delete(data, "time")
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
		delete(data, "msg")
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
		delete(data, "level")
	}

	if m, ok := data["message"]; ok {
		data["fields.message"] = m
		delete(data, "message")
	}

	if l, ok := data["timestamp"]; ok {
		data["fields.timestamp"] = l
		delete(data, "timestamp")
	}

	if l, ok := data["severity"]; ok {
		data["fields.severity"] = l
		delete(data, "severity")
	}

	if reportCaller {
		if l, ok := data[logrus.FieldKeyFunc]; ok {
			data["fields."+logrus.FieldKeyFunc] = l
		}
		if l, ok := data[logrus.FieldKeyFile]; ok {
			data["fields."+logrus.FieldKeyFile] = l
		}
	}
}
