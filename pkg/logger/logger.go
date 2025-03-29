package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Songmu/flextime"
	zlog "github.com/rs/zerolog"
)

const defaultLogLevel = zlog.InfoLevel

var (
	logger zlog.Logger
)

func Init(w io.Writer) {
	zlog.TimeFieldFormat = time.RFC3339
	zlog.TimestampFunc = flextime.Now
	zlog.LevelErrorValue = "E"
	zlog.LevelWarnValue = "W"
	zlog.LevelInfoValue = "I"
	zlog.LevelDebugValue = "D"
	zlog.MessageFieldName = "msg"
	zlog.ErrorHandler = func(err error) {
		fmt.Printf(
			`{"level": "E", "time": "%v", "msg": "%v", "error": "%v"}
`,
			flextime.Now().Format(time.RFC3339), "write log failed using zerolog", err,
		)
	}

	level := defaultLogLevel
	envLevel := os.Getenv("LOG_LEVEL")
	if envLevel != "" {
		// 先頭の文字を取得
		envLevelHead := string([]rune(strings.ToUpper(envLevel))[0])
		l, err := zlog.ParseLevel(envLevelHead)
		if err == nil && l != zlog.NoLevel {
			level = l
		}
	}
	zlog.SetGlobalLevel(level)

	logger = zlog.New(w).
		With().
		Timestamp().
		Logger()
}

func Error() *LogEvent {
	return &LogEvent{
		event: logger.Error(),
	}
}

func Warn() *LogEvent {
	return &LogEvent{
		event: logger.Warn(),
	}
}

func Info() *LogEvent {
	return &LogEvent{
		event: logger.Info(),
	}
}

func Debug() *LogEvent {
	return &LogEvent{
		event: logger.Debug(),
	}
}

type LogEvent struct {
	event *zlog.Event
}

func (l *LogEvent) Msg(msg string) {
	l.event.Msg(msg)
}

func (l *LogEvent) Msgf(format string, v ...any) {
	l.event.Msgf(format, v...)
}
