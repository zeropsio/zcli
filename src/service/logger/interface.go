package logger

type Logger interface {
	Info(a ...interface{})
	Warning(a ...interface{})
	Error(a ...interface{})
	Debug(a ...interface{})
}
