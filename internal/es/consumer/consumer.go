package consumer

import (
	"context"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkago "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	Topic string

	consumer *kafkago.Consumer
}

type Config struct {
	Address string
	Topic   string
}

func New(consumer *kafka.Consumer, config Config) *Consumer {
	err := consumer.Subscribe(config.Topic, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &Consumer{
		consumer: consumer,
		Topic:    config.Topic,
	}
}

func (cons *Consumer) ReadMessage(ctx context.Context, topic string) (*kafka.Message, error) {
	ev := cons.consumer.Poll(100)
	switch e := ev.(type) {
	case *kafka.Message:
		fmt.Printf("processing order: %s\n", string(e.Value))
		return ev.(*kafka.Message), nil
	case *kafka.Error:
		return nil, fmt.Errorf("%v\n", e)
	}
	return nil, nil
}
