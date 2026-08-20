package logger

type Logger interface {
	Debug(message string, extras ...any)
	Info(message string, extras ...any)
	Warn(message string, extras ...any)
	Error(message string, extras ...any)
}
