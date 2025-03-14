package messaging

import "context"

type (
	// Publisher is the message publishing component.
	Publisher interface {
		// Publish publishes provided messages to the given topic.
		Publish(topic string, messages ...*Message) error

		// Close should flush unsent messages if publisher is async.
		Close() error
	}

	// Subscriber is the message subscribing component.
	Subscriber interface {
		// Subscribe returns an output channel with messages from the provided topic.
		// The channel is closed after Close() is called on the subscriber.
		Subscribe(ctx context.Context, topic string) (<-chan Message, error)

		// Close closes all subscriptions with their output channels.
		Close() error
	}
)
