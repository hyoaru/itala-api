package logger

var defaultLogger Logger = NewSlogLogger()

func Debug(message string, extras ...map[string]any) {
	defaultLogger.Debug(message, extras...)
}

func Info(message string, extras ...map[string]any) {
	defaultLogger.Info(message, extras...)
}

func Warn(message string, extras ...map[string]any) {
	defaultLogger.Warn(message, extras...)
}

func Error(message string, extras ...map[string]any) {
	defaultLogger.Error(message, extras...)
}
