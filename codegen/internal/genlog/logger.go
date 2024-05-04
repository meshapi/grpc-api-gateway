package genlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Level uint8

const (
	LevelWarning Level = iota
	LevelInfo
	LevelTrace
	LevelSilent
)

func (l Level) String() string {
	switch l {
	case LevelWarning:
		return "warning"
	case LevelInfo:
		return "info"
	case LevelTrace:
		return "trace"
	default:
		return "silent"
	}
}

func (l *Level) Set(value string) error {
	switch strings.ToLower(value) {
	case "warning":
		*l = LevelWarning
	case "info":
		*l = LevelInfo
	case "trace":
		*l = LevelTrace
	case "silent":
		*l = LevelSilent
	default:
		return fmt.Errorf("unrecognized level %q, accepted values are 'warning', 'info', 'trace' and 'silent'", value)
	}

	return nil
}

// type Logger struct {Logger describes ability to
type Logger struct {
	level  Level
	logger *log.Logger
}

func New(output io.Writer, level Level) Logger {
	return Logger{level: level, logger: log.New(output, "", 0)}
}

// Log writes a log record.
func (l Logger) Log(level Level, pattern string, args ...any) {
	if l.level < level {
		return
	}

	l.logger.Printf(pattern, args...)
}

func (l Logger) HasLevel(level Level) bool {
	return l.level >= level
}

var (
	defaultLogger = Logger{
		level:  LevelWarning,
		logger: log.New(os.Stderr, "", 0),
	}
)

func Set(logger Logger) {
	defaultLogger = logger
}

func Log(level Level, pattern string, args ...any) {
	defaultLogger.Log(level, pattern, args...)
}

func HasLevel(level Level) bool {
	return defaultLogger.HasLevel(level)
}

func SetLevel(level Level) {
	defaultLogger.level = level
}
