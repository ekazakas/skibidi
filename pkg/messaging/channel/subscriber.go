package channel

import (
	"context"
	"github.com/ekazakas/skibidi/pkg/log"
	"github.com/ekazakas/skibidi/pkg/messaging"
	"sync"
)

type (
	subscriber struct {
		ctx     context.Context
		sending sync.Mutex
		output  chan messaging.Message
		logger  log.Adapter
		closed  bool
		closing chan struct{}
	}
)

func (s *subscriber) Send(message messaging.Message, logFields log.Fields) {
	s.sending.Lock()
	defer s.sending.Unlock()

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	for {
		s.logger.Trace("Sending msg to subscriber channel", logFields)

		if s.closed {
			s.logger.Info("Subscriber is closed, discarding message", logFields)

			return
		}

		// copy the message to prevent ack/nack propagation to other consumers
		// also allows to make retries on a fresh copy of the original message
		msg := messaging.NewMessage(ctx, message.Headers(), message.Body())

		select {
		case <-s.closing:
			s.logger.Trace("Closing, message discarded", logFields)

			return
		case s.output <- msg:
			s.logger.Trace("Sent message to subscriber channel", logFields)
		}

		select {
		case <-s.closing:
			s.logger.Trace("Closing, message discarded", logFields)

			return
		case <-msg.Acked():
			s.logger.Trace("Ack received", logFields)

			return
		case <-msg.Nacked():
			s.logger.Trace("Nack received, resending message", logFields)
		}
	}
}

func (s *subscriber) Close() {
	if s.closed {
		return
	}
	close(s.closing)

	s.logger.Debug("Closing subscriber, waiting for sending Lock", nil)

	// ensuring that we are not sending to closed channel
	s.sending.Lock()
	defer s.sending.Unlock()

	s.logger.Debug("Subscriber closed", nil)
	s.closed = true

	close(s.output)
}
