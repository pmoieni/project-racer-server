package lib

import (
	"log/slog"
	"os"
)

type Logger struct {
	name string
	slog *slog.Logger
}

func NewLogger(name string) *Logger {
	return &Logger{
		name: name,
		slog: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (l *Logger) Info(msg string) {
	l.slog.Info(msg, "service", l.name)
}

func (l *Logger) Warn(msg string) {
	l.slog.Warn(msg, "service", l.name)
}

func (l *Logger) Error(msg string) {
	l.slog.Error(msg, "service", l.name)
}
