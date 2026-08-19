package logger

type Logger interface {
	Debug(message string, extras ...map[string]any)
	Info(message string, extras ...map[string]any)
	Warn(message string, extras ...map[string]any)
	Error(message string, extras ...map[string]any)
}
