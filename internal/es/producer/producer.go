package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkago "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	Topic string

	producer   *kafkago.Producer
	deliveryCh chan kafka.Event
}

type Config struct {
	Address string
	Topic   string
}

func New(producer *kafka.Producer, config Config) *Producer {
	return &Producer{
		producer:   producer,
		Topic:      config.Topic,
		deliveryCh: make(chan kafka.Event, 10000),
	}
}

func (w *Producer) WriteMessages(ctx context.Context, topic string, messages ...[]byte) error {
	for _, message := range messages {
		var (
			format  = fmt.Sprintf(" - %s ", string(message))
			payload = []byte(format)
		)

		err := w.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: payload,
		},
			w.deliveryCh,
		)
		if err != nil {
			log.Fatal(err)
			return err
		}

		<-w.deliveryCh
	}
	return nil
}
