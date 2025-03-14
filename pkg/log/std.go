package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type (
	// StdAdapter is a logger implementation, which sends all logs to provided standard output.
	StdAdapter struct {
		errorLogger *log.Logger
		infoLogger  *log.Logger
		debugLogger *log.Logger
		traceLogger *log.Logger
		fields      Fields
	}
)

// NewStdAdapter creates StdAdapter which sends all logs to stderr.
func NewStdAdapter(debug, trace bool) Adapter {
	return NewStdAdapterWithOut(os.Stderr, debug, trace)
}

// NewStdAdapterWithOut creates StdAdapter which sends all logs to provided io.Writer.
func NewStdAdapterWithOut(out io.Writer, debug bool, trace bool) Adapter {
	l := log.New(out, "[skibidi] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	adapter := &StdAdapter{
		infoLogger:  l,
		errorLogger: l,
	}

	if debug {
		adapter.debugLogger = l
	}

	if trace {
		adapter.traceLogger = l
	}

	return adapter
}

func (l *StdAdapter) Error(msg string, err error, fields Fields) {
	if fields == nil {
		fields = make(Fields)
	}

	l.log(l.errorLogger, "ERROR", msg, fields.Add(Fields{"err": err}))
}

func (l *StdAdapter) Info(msg string, fields Fields) {
	l.log(l.infoLogger, "INFO", msg, fields)
}

func (l *StdAdapter) Debug(msg string, fields Fields) {
	l.log(l.debugLogger, "DEBUG", msg, fields)
}

func (l *StdAdapter) Trace(msg string, fields Fields) {
	l.log(l.traceLogger, "TRACE", msg, fields)
}

func (l *StdAdapter) With(fields Fields) Adapter {
	return &StdAdapter{
		errorLogger: l.errorLogger,
		infoLogger:  l.infoLogger,
		debugLogger: l.debugLogger,
		traceLogger: l.traceLogger,
		fields:      l.fields.Add(fields),
	}
}

func (l *StdAdapter) log(logger *log.Logger, level string, msg string, fields Fields) {
	if logger == nil {
		return
	}

	finalFields := l.fields.Add(fields)

	items := make([]string, 0, len(finalFields))
	for key, value := range finalFields {
		items = append(items, fmt.Sprintf("%s=%v", key, value))
	}
	items = append(items, fmt.Sprintf("level=%s", level), fmt.Sprintf("msg=%s", msg))

	err := logger.Output(3, fmt.Sprintf("\t%s", strings.Join(items, " ")))
	if err != nil {
		panic(err)
	}
}
