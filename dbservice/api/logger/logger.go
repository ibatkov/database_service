package logger

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Stub struct {
	InfoStub  func(args ...interface{})
	ErrorStub func(args ...interface{})
}

func (stub Stub) Info(args ...interface{}) {
	if stub.InfoStub != nil {
		stub.InfoStub(args...)
		return
	}
	panic("No implementation for 'Info' found")
}

func (stub Stub) Error(args ...interface{}) {
	if stub.ErrorStub != nil {
		stub.ErrorStub(args...)
		return
	}
	panic("No implementation for 'Error' found")
}
