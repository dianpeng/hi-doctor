package trace

import (
	"fmt"
	"log"
)

// For task tracing purpose. Currently the logging is all the around and we
// want to make the trace/log customizable

type DescriptorContext interface {
	TracePrefix() string
}

type Trace interface {
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

type trace struct {
	prefix DescriptorContext
}

func (t *trace) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("INFO [%s]: %s\n", t.prefix.TracePrefix(), msg)
}

func (t *trace) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("WARN [%s]: %s\n", t.prefix.TracePrefix(), msg)
}

func (t *trace) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("ERROR [%s]: %s\n", t.prefix.TracePrefix(), msg)
}

func NewTrace(x DescriptorContext) Trace {
	return &trace{
		prefix: x,
	}
}
