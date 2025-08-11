package observability

import (
    "go.uber.org/zap"
)

type Logger struct {
    *zap.SugaredLogger
}

func NewLogger(cfg interface{}) *Logger {
    // simple dev logger; could switch based on cfg in the future
    l, _ := zap.NewDevelopment()
    return &Logger{l.Sugar()}
}

func (l *Logger) Sync() error { return l.SugaredLogger.Sync() }

// Field helpers to avoid leaking zap in other packages
func Field(key string, value interface{}) interface{} { return zap.Any(key, value) }
func Error(err error) interface{} { return zap.Error(err) }

