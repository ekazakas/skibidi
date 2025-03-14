package log_test

import (
	"errors"
	"github.com/ekazakas/skibidi/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

type (
	successfulWriter struct {
		bytesWritten [][]byte
	}

	failingWriter struct {
		err error
	}
)

func newSuccessfulWriter() *successfulWriter {
	return &successfulWriter{
		bytesWritten: [][]byte{},
	}
}

func (w *successfulWriter) Write(p []byte) (int, error) {
	current := make([]byte, len(p))
	copy(current, p)

	w.bytesWritten = append(w.bytesWritten, current)

	return len(p), nil
}

func newFailingWriter(err error) *failingWriter {
	return &failingWriter{
		err: err,
	}
}

func (w *failingWriter) Write(p []byte) (int, error) {
	return len(p), w.err
}

func TestStdAdapter_DisabledLevels(t *testing.T) {
	writer := newSuccessfulWriter()
	adapter := log.NewStdAdapterWithOut(writer, false, false).With(log.Fields{"init": "global"})

	expectedErr := errors.New("fail")

	adapter.Info("test", log.Fields{"foo": "bar"})
	adapter.Error("test", expectedErr, log.Fields{"foo": "bar"})
	adapter.Debug("test", log.Fields{"foo": "bar"})
	adapter.Trace("test", log.Fields{"foo": "bar"})

	assert.Len(t, writer.bytesWritten, 2)
	assert.Contains(t, string(writer.bytesWritten[0]), "init=global foo=bar level=INFO msg=test")
	assert.Contains(t, string(writer.bytesWritten[1]), "init=global foo=bar err=fail level=ERROR msg=test")
}

func TestStdAdapter_EnabledLevels(t *testing.T) {
	writer := newSuccessfulWriter()
	adapter := log.NewStdAdapterWithOut(writer, true, true).With(log.Fields{"init": "global"})

	expectedErr := errors.New("fail")

	adapter.Info("test", log.Fields{"foo": "bar"})
	adapter.Error("test", expectedErr, log.Fields{"foo": "bar"})
	adapter.Debug("test", log.Fields{"foo": "bar"})
	adapter.Trace("test", log.Fields{"foo": "bar"})

	assert.Len(t, writer.bytesWritten, 4)
	assert.Contains(t, string(writer.bytesWritten[0]), "init=global foo=bar level=INFO msg=test")
	assert.Contains(t, string(writer.bytesWritten[1]), "init=global foo=bar err=fail level=ERROR msg=test")
	assert.Contains(t, string(writer.bytesWritten[2]), "init=global foo=bar level=DEBUG msg=test")
	assert.Contains(t, string(writer.bytesWritten[3]), "init=global foo=bar level=TRACE msg=test")
}

func TestStdAdapter_FailingWriter(t *testing.T) {
	expectedErr := errors.New("fail")

	adapter := log.NewStdAdapterWithOut(newFailingWriter(expectedErr), true, true)

	assert.PanicsWithError(
		t,
		expectedErr.Error(),
		func() {
			adapter.Info("test", log.Fields{"foo": "bar"})
		},
	)
}
