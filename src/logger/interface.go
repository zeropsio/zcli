package logger

type Logger interface {
	Info(...interface{})
	Infof(string, ...interface{})
	Warning(a ...interface{})
	Warningf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
}
