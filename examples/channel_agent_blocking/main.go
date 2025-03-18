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
			Blocking: true,
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

	results, err := agent.Subscribe(context.Background(), topic)
	if err != nil {
		panic(err)
	}

	messages := []messaging.Message{
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-1")),
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-2")),
		messaging.NewMessage(context.Background(), messaging.Headers{}, messaging.Body("test-3")),
	}

	go receive(len(messages), results)

	for i, msg := range messages {
		fmt.Printf("Sending message #%d - %+v\n", i, msg)

		if err := agent.Publish(topic, msg); err != nil {
			panic(err)
		}
	}
}

func receive(amount int, results <-chan messaging.Message) {
	for i := 0; i < amount; i++ {
		fmt.Printf("Revceiving message #%d\n", i)
		result := <-results

		fmt.Printf("received: %+v\n", result)

		result.Ack()
	}
}
