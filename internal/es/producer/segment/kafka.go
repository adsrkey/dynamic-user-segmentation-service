package segment

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
	Value domain.Segment
}

func New(producer *producer.Producer, topicName string) *Producer {
	return &Producer{
		producer:  producer,
		topicName: topicName,
	}
}

func (w *Producer) PublishCreated(ctx context.Context, segment domain.Segment) error {
	return w.publish(ctx, "segment.event.created", segment)
}

func (w *Producer) PublishDeleted(ctx context.Context, segment domain.Segment) error {
	return w.publish(ctx, "segment.event.deleted", segment)
}

func (w *Producer) publish(ctx context.Context, msgType string, segment domain.Segment) error {
	var b bytes.Buffer

	evt := Event{
		Type:  msgType,
		Value: segment,
	}

	if err := json.NewEncoder(&b).Encode(evt); err != nil {
		return err
	}

	return w.producer.WriteMessages(ctx, w.topicName, b.Bytes())
}
