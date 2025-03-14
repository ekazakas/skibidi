package log

type (
	// NoopAdapter is a logger which discards all logs.
	NoopAdapter struct{}
)

func (NoopAdapter) Error(string, error, Fields) {}

func (NoopAdapter) Info(string, Fields) {}

func (NoopAdapter) Debug(string, Fields) {}

func (NoopAdapter) Trace(string, Fields) {}

func (l NoopAdapter) With(Fields) Adapter {
	return l
}
