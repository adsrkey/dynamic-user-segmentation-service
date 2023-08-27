package user

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

func (w *Consumer) ReadUserAddToSegment(ctx context.Context, user *domain.User) error {
	return w.read(ctx, "user.event.add_to_segment", user)
}

func (w *Consumer) read(ctx context.Context, msgType string, user *domain.User) error {
	message, err := w.consumer.ReadMessage(ctx, w.topicName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(message.Value, user)
	if err != nil {
		return err
	}

	return nil
}
