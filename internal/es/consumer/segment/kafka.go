package segment

import (
	"context"
	"encoding/json"

	"github.com/adsrkey/dynamic-user-segmentation-service/internal/domain"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/es/consumer"
)

type Consumer struct {
	consumer  *consumer.Consumer
	topicName string
}

type Event struct {
	Type  string
	Value domain.Segment
}

func New(consumer *consumer.Consumer, topicName string) *Consumer {
	return &Consumer{
		consumer:  consumer,
		topicName: topicName,
	}
}

func (w *Consumer) ReadCreated(ctx context.Context, segment *domain.Segment) error {
	return w.read(ctx, "segment.event.created", segment)
}

func (w *Consumer) ReadDeleted(ctx context.Context, segment *domain.Segment) error {
	return w.read(ctx, "segment.event.deleted", segment)
}

func (w *Consumer) read(ctx context.Context, msgType string, segment *domain.Segment) error {
	message, err := w.consumer.ReadMessage(ctx, w.topicName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(message.Value, segment)
	if err != nil {
		return err
	}

	return nil
}
