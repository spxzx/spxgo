package log

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
	TimeDisplay bool
}

func (f *JsonFormatter) Format(param *LoggerFormatParam) string {
	now := time.Now()
	if param.LoggerFields == nil {
		param.LoggerFields = make(Fields)
	}
	if f.TimeDisplay {
		param.LoggerFields["log_time"] = now.Format("2006/01/02 - 15:04:05")
	}
	param.LoggerFields["message"] = param.Message
	param.LoggerFields["log_level"] = param.Level.Level()
	marshal, err := json.Marshal(param.LoggerFields)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", marshal)
}
