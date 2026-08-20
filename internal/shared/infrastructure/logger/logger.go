package logger

var defaultLogger Logger = NewSlogLogger()

func Debug(message string, extras ...any) {
	defaultLogger.Debug(message, extras...)
}

func Info(message string, extras ...any) {
	defaultLogger.Info(message, extras...)
}

func Warn(message string, extras ...any) {
	defaultLogger.Warn(message, extras...)
}

func Error(message string, extras ...any) {
	defaultLogger.Error(message, extras...)
}
