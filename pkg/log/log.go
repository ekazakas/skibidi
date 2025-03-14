package log

const (
	TraceLevel Level = iota + 1
	DebugLevel
	InfoLevel
	ErrorLevel
)

type (
	// Fields is the logger's key-value list of fields.
	Fields map[string]interface{}

	// Adapter is an interface, that you need to implement to support logging.
	// You can use StdAdapter as a reference implementation.
	Adapter interface {
		Error(msg string, err error, fields Fields)
		Info(msg string, fields Fields)
		Debug(msg string, fields Fields)
		Trace(msg string, fields Fields)
		With(fields Fields) Adapter
	}

	Level uint
)

// Add adds new fields to the list of Fields.
func (l Fields) Add(newFields Fields) Fields {
	resultFields := make(Fields, len(l)+len(newFields))

	for field, value := range l {
		resultFields[field] = value
	}

	for field, value := range newFields {
		resultFields[field] = value
	}

	return resultFields
}

// Copy copies the Fields.
func (l Fields) Copy() Fields {
	fieldsCopy := make(Fields, len(l))

	for k, v := range l {
		fieldsCopy[k] = v
	}

	return fieldsCopy
}
