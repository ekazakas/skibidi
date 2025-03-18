package channel

import (
	"context"
	"errors"
	"fmt"
	"github.com/ekazakas/skibidi/pkg/log"
	"github.com/ekazakas/skibidi/pkg/messaging"
	"github.com/google/uuid"
	"math/rand"
	"sync"
)

type (
	// Config holds the Agent configuration options.
	Config struct {
		// Size output channel buffer size.
		Size int64

		// Persistent when is set to true, when subscriber subscribes to the topic,
		// it will receive all previously produced messages.
		//
		// All messages are persisted to the memory.
		Persistent bool

		// Blocking when true, Publish will bLock until subscriber Ack's the message.
		// If there are no subscribers, Publish will not bLock.
		Blocking bool
	}

	// Agent is the simplest Pub/Sub implementation.
	// It is based on Golang's channels which are sent within the process.
	//
	// Agent has no global state,
	// that means that you need to use the same instance for Publishing and Subscribing!
	//
	// When Agent is persistent, messages order is not guaranteed.
	Agent struct {
		config                Config
		logger                log.Adapter
		subscribersLock       sync.RWMutex
		subscribers           map[string][]*subscriber
		closedLock            sync.Mutex
		closed                bool
		closing               chan struct{}
		persistedMessagesLock sync.RWMutex
		persistedMessages     map[string][]messaging.Message
	}
)

func (c *Config) validate() error {
	if c.Size < 0 {
		return fmt.Errorf("invalid size: %d", c.Size)
	}

	return nil
}

// NewAgent creates new Agent Pub/Sub.
//
// This Agent is not persistent.
// That means if you send a message to a topic to which no subscriber is subscribed, that message will be discarded.
func NewAgent(config Config) (*Agent, error) {
	return NewAgentWithLogger(config, log.NoopAdapter{})
}

func NewAgentWithLogger(config Config, logger log.Adapter) (*Agent, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	if logger == nil {
		return nil, errors.New("logger must not be nil")
	}

	return &Agent{
		config:            config,
		logger:            logger.With(log.Fields{"agent_id": uuid.New().String()}),
		subscribers:       make(map[string][]*subscriber),
		closing:           make(chan struct{}),
		persistedMessages: map[string][]messaging.Message{},
	}, nil
}

// Publish in Agent is NOT bLocking until all consumers consume.
// Messages will be sent in background.
//
// Messages may be persisted or not, depending on persistent attribute.
func (c *Agent) Publish(topic string, messages ...messaging.Message) error {
	if c.isClosed() {
		return errors.New("channel is closed")
	}

	if c.config.Persistent {
		c.persistMessages(topic, messages)
	}

	for _, message := range messages {
		logFields := log.Fields{
			"topic": topic,
		}

		completed, err := c.sendMessage(topic, message, logFields)
		if err != nil {
			return err
		}

		if c.config.Blocking {
			c.waitForSubscribers(logFields, completed)
		}
	}

	return nil
}

// Subscribe returns channel to which all published messages are sent.
// Messages are not persisted. If there are no subscribers and message is produced it will be gone.
//
// There are no consumer groups support etc. Every consumer will receive every produced message.
func (c *Agent) Subscribe(ctx context.Context, topic string) (<-chan messaging.Message, error) {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()

	if c.closed {
		return nil, errors.New("agent closed")
	}
	//
	//c.subscribersWc.Add(1)
	//c.closedLock.Unlock()
	//
	//c.subscribersLock.Lock()
	//
	//subLock, _ := c.subscribersByTopicLock.LoadOrStore(topic, &sync.Mutex{})
	//subLock.(*sync.Mutex).Lock()
	//
	sub := &subscriber{
		ctx:    ctx,
		output: make(chan messaging.Message, c.config.Size),
		logger: c.logger.With(log.Fields{
			"subscriber_id": rand.Uint64(),
		}),
		closing: make(chan struct{}),
	}

	go func(s *subscriber, c *Agent) {
		select {
		case <-ctx.Done():
		case <-c.closing:
		}

		s.Close()
	}(sub, c)

	if !c.config.Persistent {
		c.addSubscriber(topic, sub)

		return sub.output, nil
	}

	go func(s *subscriber) {
		if messages := c.fetchMessages(topic); len(messages) > 0 {
			for _, message := range messages {
				go s.Send(message, log.Fields{
					"topic": topic,
				})
			}
		}

		c.addSubscriber(topic, s)
	}(sub)

	return sub.output, nil
}

// Close closes the Agent.
func (c *Agent) Close() error {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closing)

	return nil
}

func (c *Agent) persistMessages(topic string, messages []messaging.Message) {
	c.persistedMessagesLock.Lock()
	defer c.persistedMessagesLock.Unlock()

	if _, ok := c.persistedMessages[topic]; !ok {
		c.persistedMessages[topic] = make([]messaging.Message, len(messages))
	}

	c.persistedMessages[topic] = append(c.persistedMessages[topic], messages...)
}

func (c *Agent) fetchMessages(topic string) []messaging.Message {
	c.persistedMessagesLock.RLock()
	defer c.persistedMessagesLock.RUnlock()

	if messages, ok := c.persistedMessages[topic]; ok {
		return messages
	}

	return nil
}

func (c *Agent) isClosed() bool {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()

	return c.closed
}

func (c *Agent) addSubscriber(topic string, sub *subscriber) {
	c.subscribersLock.Lock()
	defer c.subscribersLock.Unlock()

	if _, ok := c.subscribers[topic]; !ok {
		c.subscribers[topic] = make([]*subscriber, 0)
	}

	c.subscribers[topic] = append(c.subscribers[topic], sub)
}

func (c *Agent) topicSubscribers(topic string) []*subscriber {
	c.subscribersLock.RLock()
	defer c.subscribersLock.RUnlock()

	subscribers, ok := c.subscribers[topic]
	if !ok {
		return nil
	}

	return subscribers
}

func (c *Agent) sendMessage(topic string, message messaging.Message, logFields log.Fields) (<-chan struct{}, error) {
	subscribers := c.topicSubscribers(topic)
	completed := make(chan struct{})

	if len(subscribers) == 0 {
		close(completed)
		c.logger.Info("No subscribers to send message", logFields)

		return completed, nil
	}

	go func(subscribers []*subscriber) {
		wg := &sync.WaitGroup{}
		wg.Add(len(subscribers))

		for _, sub := range subscribers {
			go func() {
				sub.Send(message, logFields)
				wg.Done()
			}()
		}

		wg.Wait()
		close(completed)
	}(subscribers)

	return completed, nil
}

func (c *Agent) waitForSubscribers(logFields log.Fields, completed <-chan struct{}) {
	c.logger.Debug("Waiting for subscribers", logFields)

	select {
	case <-c.closing:
		c.logger.Trace("Closing agent before completion from subscribers", logFields)
	case <-completed:
		c.logger.Trace("Subscribers completed", logFields)
	}
}
