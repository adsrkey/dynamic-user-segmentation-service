package user

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/adsrkey/dynamic-user-segmentation-service/internal/domain"
	producer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/producer"
)

type Producer struct {
	producer  *producer.Producer
	topicName string
}

type Event struct {
	Type  string
	Value domain.User
}

func New(producer *producer.Producer, topicName string) *Producer {
	return &Producer{
		producer:  producer,
		topicName: topicName,
	}
}

func (w *Producer) PublishUserAddToSegment(ctx context.Context, user domain.User) error {
	return w.publish(ctx, "user.event.add_to_segment", user)
}

func (w *Producer) publish(ctx context.Context, msgType string, user domain.User) error {
	var b bytes.Buffer

	evt := Event{
		Type:  msgType,
		Value: user,
	}

	if err := json.NewEncoder(&b).Encode(evt); err != nil {
		return err
	}

	return w.producer.WriteMessages(ctx, w.topicName, b.Bytes())
}
