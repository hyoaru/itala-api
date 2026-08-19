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

func (l *SlogLogger) Debug(message string, extras ...map[string]any) {
	if len(extras) == 0 {
		l.logger.Debug(message)
		return
	}

	l.logger.Debug(message, "extras", extras[0])
}

func (l *SlogLogger) Info(message string, extras ...map[string]any) {
	if len(extras) == 0 {
		l.logger.Info(message)
		return
	}

	l.logger.Info(message, "extras", extras[0])
}

func (l *SlogLogger) Warn(message string, extras ...map[string]any) {
	if len(extras) == 0 {
		l.logger.Warn(message)
		return
	}

	l.logger.Warn(message, "extras", extras[0])
}

func (l *SlogLogger) Error(message string, extras ...map[string]any) {
	if len(extras) == 0 {
		l.logger.Error(message)
		return
	}

	l.logger.Error(message, "extras", extras[0])
}
