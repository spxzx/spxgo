package log

import (
	"fmt"
	"gitbuh.com/spxzx/spxgo/internal/sstrings"
	"io"
	"log"
	"os"
	"path"
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

type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type Fields map[string]any

type LoggerFormatter interface {
	Format(*LoggerFormatParam) string
}

type LoggerFormatParam struct {
	Message      any
	IsColor      bool
	Level        LoggerLevel
	LoggerFields Fields
}

type Logger struct {
	Formatter    LoggerFormatter
	Level        LoggerLevel
	Outs         []*LoggerWriter
	LoggerFields Fields
	logPath      string
	LogFileSize  int64
}

type LoggerWriter struct {
	Level LoggerLevel
	Out   io.Writer
}

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Level = LevelDebug
	w := &LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
	logger.Formatter = &TextFormatter{}
	return logger
}

func (l *Logger) WithFields(fields Fields) *Logger {
	l.LoggerFields = fields
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		Level:        l.Level,
		LoggerFields: l.LoggerFields,
	}
}

func (l *Logger) Print(level LoggerLevel, msg any) {
	if l.Level > level {
		// 当前级别大于输入级别 不打印对应级别日志
		return
	}
	param := &LoggerFormatParam{
		Message:      msg,
		Level:        level,
		LoggerFields: l.LoggerFields,
	}
	str := l.Formatter.Format(param)
	for _, w := range l.Outs {
		if w.Out == os.Stdout {
			param.IsColor = true
			temp := str
			str = l.Formatter.Format(param)
			_, _ = fmt.Fprintln(w.Out, str)
			str = temp
		} else if w.Level == -1 || w.Level == level {
			_, _ = fmt.Fprintln(w.Out, str)
			l.CheckFileSize(w)
		}
	}
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}

func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644) // 0644 文件权限
	if err != nil {
		panic(err)
	}
	return w
}

func (l *Logger) SetLogPath(logPath string) {
	l.logPath = logPath
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: -1,
		Out:   FileWriter(path.Join(logPath, "log.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: 0,
		Out:   FileWriter(path.Join(logPath, "debug.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: 1,
		Out:   FileWriter(path.Join(logPath, "info.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: 2,
		Out:   FileWriter(path.Join(logPath, "error.log")),
	})
}

func (l *Logger) CheckFileSize(w *LoggerWriter) {
	logFile := w.Out.(*os.File)
	if logFile != nil {
		stat, err := logFile.Stat()
		if err != nil {
			log.Panicln(err)
			return
		}
		size := stat.Size()
		if l.LogFileSize <= 0 {
			l.LogFileSize = 32 << 10
		}
		if size >= l.LogFileSize {
			_, name := path.Split(stat.Name())
			fileName := name[0:strings.Index(name, ".")]
			if strings.Index(fileName, "_") != -1 {
				fileName = fileName[0:strings.Index(fileName, "_")]
			}
			writer := FileWriter(path.Join(l.logPath,
				sstrings.JoinStrings(fileName, "_", time.Now().Format("2006_01_02_15_04_05"),
					".log")))
			w.Out = writer
		}
	}
}

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

func (f *LoggerFormatParam) DefaultLevelColor() string {
	switch f.Level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return ""
	}
}

func (f *LoggerFormatParam) DefaultMsgColor() string {
	switch f.Level {
	case LevelError:
		return red
	default:
		return ""
	}
}
