package spxgo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

type LoggerFormatter = func(params *LogFormatterParams) string

var defaultLogFormatter = func(params *LogFormatterParams) string {
	statusCodeColor := params.StatusCodeColor()
	reset := params.ResetColor()
	if params.Latency > time.Minute { // 执行时间超过一分钟显示成秒
		params.Latency = params.Latency.Truncate(time.Second)
	}
	if params.IsDisplayColor {
		return fmt.Sprintf("[spx] %v |%s %3d %s| %12v | %15s | %-s | %#v",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusCodeColor, params.StatusCode, reset,
			params.Latency, params.ClientIP, params.Method, params.Path,
		)
	}
	return fmt.Sprintf("[spx] %v | %3d | %12v | %15s | %-s | %#v",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		params.StatusCode,
		params.Latency, params.ClientIP, params.Method, params.Path,
	)
}

var defaultWriter io.Writer = os.Stdout

func LoggingWithConfig(conf LoggingConfig, next HandlerFunc) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}
	out := conf.out
	if out == nil {
		out = defaultWriter
	}
	return func(c *Context) {
		r := c.R
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery
		next(c)
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := r.Method
		statusCode := c.StatusCode
		if raw != "" {
			path = path + "?" + raw
		}
		param := &LogFormatterParams{
			Request:        c.R,
			TimeStamp:      stop,
			StatusCode:     statusCode,
			Latency:        latency,
			ClientIP:       clientIP,
			Method:         method,
			Path:           path,
			IsDisplayColor: true,
		}
		_, _ = fmt.Fprintln(out, formatter(param))
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
