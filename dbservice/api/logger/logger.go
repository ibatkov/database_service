package logger

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}
