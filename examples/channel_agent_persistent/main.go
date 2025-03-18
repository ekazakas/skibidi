package main

import (
	"context"
	"fmt"
	"github.com/ekazakas/skibidi/pkg/messaging"
	"github.com/ekazakas/skibidi/pkg/messaging/channel"
)

func main() {
	agent, err := channel.NewAgent(
		channel.Config{
			Persistent: true,
		},
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := agent.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "test-topic"

	messages := []messaging.Message{
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-1")),
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-2")),
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-3")),
	}

	for i, msg := range messages {
		fmt.Printf("Sending message #%d - %+v\n", i, msg)

		if err := agent.Publish(topic, msg); err != nil {
			panic(err)
		}
	}

	results1, err := agent.Subscribe(context.Background(), topic)
	if err != nil {
		panic(err)
	}

	results2, err := agent.Subscribe(context.Background(), topic)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(messages); i++ {
		receive(i, results1)
		receive(i, results2)
	}
}

func receive(i int, results <-chan messaging.Message) {
	fmt.Printf("Revceiving message #%d\n", i)
	result := <-results

	fmt.Printf("received: %+v\n", result)

	result.Ack()
}
