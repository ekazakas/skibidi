package log

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

type (
	// CaptureAdapter is a logger which captures all logs.
	// This logger is mostly useful for testing logging.
	CaptureAdapter struct {
		captured map[Level][]CapturedMessage
		fields   Fields
		lock     *sync.Mutex
	}

	CapturedMessage struct {
		Level  Level
		Time   time.Time
		Fields Fields
		Msg    string
		Err    error
	}
)

func (c CapturedMessage) ContentEquals(other CapturedMessage) bool {
	return c.Level == other.Level &&
		reflect.DeepEqual(c.Fields, other.Fields) &&
		c.Msg == other.Msg &&
		errors.Is(c.Err, other.Err)
}

func NewCaptureAdapter() *CaptureAdapter {
	return &CaptureAdapter{
		captured: map[Level][]CapturedMessage{},
		lock:     &sync.Mutex{},
	}
}

func (c *CaptureAdapter) With(fields Fields) Adapter {
	c.lock.Lock()
	defer c.lock.Unlock()

	return &CaptureAdapter{
		captured: c.captured, // we are passing the same map, so we'll capture logs from this instance as well
		fields:   c.fields.Copy().Add(fields),
		lock:     c.lock,
	}
}

func (c *CaptureAdapter) capture(level Level, msg string, err error, fields Fields) {
	c.lock.Lock()
	defer c.lock.Unlock()

	logMsg := CapturedMessage{
		Level:  level,
		Time:   time.Now(),
		Fields: c.fields.Add(fields),
		Msg:    msg,
		Err:    err,
	}

	c.captured[level] = append(c.captured[level], logMsg)
}

func (c *CaptureAdapter) Captured() map[Level][]CapturedMessage {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.captured
}

func (c *CaptureAdapter) Has(msg CapturedMessage) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, capturedMsg := range c.captured[msg.Level] {
		if msg.ContentEquals(capturedMsg) {
			return true
		}
	}

	return false
}

func (c *CaptureAdapter) HasError(err error) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, capturedMsg := range c.captured[ErrorLevel] {
		if errors.Is(err, capturedMsg.Err) {
			return true
		}
	}
	return false
}

func (c *CaptureAdapter) Error(msg string, err error, fields Fields) {
	c.capture(ErrorLevel, msg, err, fields)
}

func (c *CaptureAdapter) Info(msg string, fields Fields) {
	c.capture(InfoLevel, msg, nil, fields)
}

func (c *CaptureAdapter) Debug(msg string, fields Fields) {
	c.capture(DebugLevel, msg, nil, fields)
}

func (c *CaptureAdapter) Trace(msg string, fields Fields) {
	c.capture(TraceLevel, msg, nil, fields)
}
