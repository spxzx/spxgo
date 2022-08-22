package log

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(param *LoggerFormatParam) string {
	now := time.Now()
	fieldsString := ""
	if param.LoggerFields != nil {
		var sb strings.Builder
		cnt, length := 0, len(param.LoggerFields)
		for k, v := range param.LoggerFields {
			_, _ = fmt.Fprintf(&sb, "%s=%v", k, v)
			if cnt < length-1 {
				_, _ = fmt.Fprintf(&sb, ", ")
				cnt++
			}
		}
		fieldsString = "- " + sb.String()
	}
	if param.IsColor {
		levelColor := param.DefaultLevelColor()
		msgColor := param.DefaultMsgColor()
		return fmt.Sprintf("[spx] %v [%s%s%s] %s%v%s %s",
			now.Format("2006/01/02 - 15:04:05"),
			levelColor, param.Level.Level(), reset,
			msgColor, param.Message, reset, fieldsString,
		)
	}
	return fmt.Sprintf("[spx] %v [%s] %v %s",
		now.Format("2006/01/02 - 15:04:05"),
		param.Level.Level(), param.Message, fieldsString,
	)
}
