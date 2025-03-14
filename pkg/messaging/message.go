package messaging

import (
	"bytes"
	"context"
	"reflect"
	"sync"
)

type (
	// Headers represent metadata associated with the Message a key-value map, providing contextual information.
	Headers map[string]string

	// Body contains the raw data payload of the Message.
	Body []byte

	// Message defines the interface for data transfer units within the system.
	// It encapsulates the message payload, metadata, and lifecycle management.
	Message interface {
		// Headers returns the message's metadata.
		// This allows access to contextual information associated with the message.
		Headers() Headers

		// Body returns the raw data payload of the message.
		Body() Body

		// Equals performs a deep comparison to determine if two Messages have identical content and metadata.
		// It returns true if the Messages are equal, false otherwise.
		Equals(other Message) bool

		// Context returns the message's context.
		// This allows for propagation of request-scoped values and cancellation signals.
		Context() context.Context

		// Ack acknowledges the successful processing of the message.
		// It returns true if the acknowledgement was successful, false otherwise.
		Ack() bool

		// Nack negatively acknowledges the processing of the message, indicating a failure.
		// It returns true if the negative acknowledgement was successful, false otherwise.
		Nack() bool

		// Acked returns a read-only channel that is closed when the message has been successfully acknowledged.
		// This allows for asynchronous notification of acknowledgement.
		Acked() <-chan struct{}

		// Nacked returns a read-only channel that is closed when the message has been negatively acknowledged.
		// This allows for asynchronous notification of negative acknowledgement.
		Nacked() <-chan struct{}
	}

	message struct {
		headers     Headers
		body        Body
		ack         chan struct{}
		nack        chan struct{}
		ackMutex    sync.Mutex
		ackSentType ackType
		ctx         context.Context
	}

	ackType int
)

const (
	noAckSent ackType = iota
	ack
	nack
)

// Get returns the value from Headers for the given key. If the key is not found, an empty string is returned.
func (h Headers) Get(key string) string {
	if v, ok := h[key]; ok {
		return v
	}

	return ""
}

// Set sets the value on Headers for a given key.
func (h Headers) Set(key, value string) {
	h[key] = value
}

// NewMessage creates a new Message with given context, headers and payload.
func NewMessage(ctx context.Context, headers Headers, body Body) Message {
	return &message{
		headers: headers,
		body:    body,
		ack:     make(chan struct{}),
		nack:    make(chan struct{}),
		ctx:     ctx,
	}
}

func (m *message) Headers() Headers {
	return m.headers
}

func (m *message) Body() Body {
	return m.body
}

func (m *message) Equals(other Message) bool {
	if len(m.Headers()) != len(other.Headers()) {
		return false
	}

	if !reflect.DeepEqual(m.Headers, other.Headers) {
		return false
	}

	return bytes.Equal(m.Body(), other.Body())
}

func (m *message) Ack() bool {
	m.ackMutex.Lock()
	defer m.ackMutex.Unlock()

	if m.ackSentType == nack {
		return false
	}

	if m.ackSentType != noAckSent {
		return true
	}

	m.ackSentType = ack
	close(m.ack)

	return true
}

func (m *message) Nack() bool {
	m.ackMutex.Lock()
	defer m.ackMutex.Unlock()

	if m.ackSentType == ack {
		return false
	}
	if m.ackSentType != noAckSent {
		return true
	}

	m.ackSentType = nack
	close(m.nack)

	return true
}

func (m *message) Acked() <-chan struct{} {
	return m.ack
}

func (m *message) Nacked() <-chan struct{} {
	return m.nack
}

func (m *message) Context() context.Context {
	return m.ctx
}
