package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Default log format will output [INFO]: 2006-01-02T15:04:05Z07:00 - Log message
	defaultLogFormat       = "[%lvl%]: %time% - %msg%"
	defaultTimestampFormat = time.RFC3339
)

// Formatter implements logrus.Formatter interface.
type Formatter struct {
	// Timestamp format
	TimestampFormat string
	OnlyTime        bool
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	LogFormat string
}

func timeFormat(t time.Time) string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}

// Format building log message.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.LogFormat
	if output == "" {
		output = defaultLogFormat
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}

	if f.OnlyTime {
		output = strings.Replace(output, "%time%", timeFormat(entry.Time), 1)
	} else {
		output = strings.Replace(output, "%time%", entry.Time.Format(timestampFormat), 1)
	}

	output = strings.Replace(output, "%msg%", entry.Message, 1)

	level := strings.ToUpper(entry.Level.String())
	output = strings.Replace(output, "%lvl%", level, 1)

	fields := ""
	for k, val := range entry.Data {
		fields += fmt.Sprintf("[%s=%s]", k, val)
		// switch v := val.(type) {
		// case string:
		// 	output = strings.Replace(output, "%fields%", v, 1)
		// case int:
		// 	s := strconv.Itoa(v)
		// 	output = strings.Replace(output, "%fields%", s, 1)
		// case bool:
		// 	s := strconv.FormatBool(v)
		// 	output = strings.Replace(output, "%fields%", s, 1)
		// }
	}
	output = strings.Replace(output, "%fields%", fields, 1)
	return []byte(output), nil
}
