package logger

import (
	"log/slog"
	"os"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger() *SlogLogger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

func (l *SlogLogger) Debug(message string, extras ...any) {
	l.logger.Debug(message, extras...)
}

func (l *SlogLogger) Info(message string, extras ...any) {
	l.logger.Info(message, extras...)
}

func (l *SlogLogger) Warn(message string, extras ...any) {
	l.logger.Warn(message, extras...)
}

func (l *SlogLogger) Error(message string, extras ...any) {
	l.logger.Error(message, extras...)
}
