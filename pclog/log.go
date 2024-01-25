package pclog

//go:generate mockgen -source=log.go -package=mock -destination=mock/log.go

import (
	"context"
	"fmt"
	"log"
	"os"
)

// Level defines the priority of a log message.
// When a logger is configured with a level, any log message with a lower
// log level (smaller by integer comparison) will not be output.
type Level int

// The levels of logs.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var levelStr = [...]string{
	"[DEBUG] ",
	"[INFO] ",
	"[WARNING] ",
	"[ERROR] ",
	"[FATAL] ",
}

func (level Level) string() string {
	return levelStr[level]
}

// CtxLogger is a logger interface that accepts a context argument and output
// logs with a format.
type CtxLogger interface {
	CtxDebugF(ctx context.Context, format string, v ...interface{})
	CtxInfoF(ctx context.Context, format string, v ...interface{})
	CtxWarnF(ctx context.Context, format string, v ...interface{})
	CtxErrorF(ctx context.Context, format string, v ...interface{})
	CtxFatalF(ctx context.Context, format string, v ...interface{})
}

const logPrefix = "polyCache "

type defaultLogger struct {
	level     Level
	stdLogger *log.Logger
}

func NewLogger(level Level) CtxLogger {
	return &defaultLogger{
		level:     level,
		stdLogger: log.New(os.Stderr, logPrefix, log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
	}
}

func (l *defaultLogger) SetLevel(level Level) {
	l.level = level
}

func (l *defaultLogger) CtxDebugF(ctx context.Context, format string, v ...interface{}) {
	l.logf(LevelDebug, format, v...)
}

func (l *defaultLogger) CtxInfoF(ctx context.Context, format string, v ...interface{}) {
	l.logf(LevelInfo, format, v...)
}

func (l *defaultLogger) CtxWarnF(ctx context.Context, format string, v ...interface{}) {
	l.logf(LevelWarning, format, v...)
}

func (l *defaultLogger) CtxErrorF(ctx context.Context, format string, v ...interface{}) {
	l.logf(LevelError, format, v...)
}

func (l *defaultLogger) CtxFatalF(ctx context.Context, format string, v ...interface{}) {
	l.logf(LevelFatal, format, v...)
}

func (l *defaultLogger) logf(level Level, format string, v ...interface{}) {
	if l.level > level {
		return
	}
	msg := level.string()
	msg += fmt.Sprintf(format, v...)
	_ = l.stdLogger.Output(3, msg)
	if level == LevelFatal {
		os.Exit(1)
	}
}
