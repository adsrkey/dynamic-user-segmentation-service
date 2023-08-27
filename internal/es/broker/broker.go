package broker

import (
	"context"

	"github.com/adsrkey/dynamic-user-segmentation-service/internal/domain"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/es/consumer"
	segmentKafkaConsumer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/consumer/segment"
	userKafkaConsumer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/consumer/user"
	producer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/producer"
	segmentKafkaProducer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/producer/segment"
	userKafkaProducer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/producer/user"
)

type Segment interface {
	SegmentProducer
	SegmentConsumer
}

type User interface {
	UserProducer
	UserConsumer
}

type UserProducer interface {
	PublishUserAddToSegment(ctx context.Context, transfer domain.User) error
}

type UserConsumer interface {
	ReadUserAddToSegment(ctx context.Context, user *domain.User) error
}

type SegmentProducer interface {
	PublishCreated(ctx context.Context, segment domain.Segment) error
	PublishDeleted(ctx context.Context, segment domain.Segment) error
}

type SegmentConsumer interface {
	ReadCreated(ctx context.Context, segment *domain.Segment) error
	ReadDeleted(ctx context.Context, segment *domain.Segment) error
}

type MessageBroker struct {
	SegmentProducer
	SegmentConsumer
	UserProducer
	UserConsumer
}

func NewKafkaMessageBroker(producer *producer.Producer, consumer *consumer.Consumer) *MessageBroker {
	return &MessageBroker{
		SegmentProducer: segmentKafkaProducer.New(producer, producer.Topic),
		SegmentConsumer: segmentKafkaConsumer.New(consumer, consumer.Topic),
		UserProducer:    userKafkaProducer.New(producer, producer.Topic),
		UserConsumer:    userKafkaConsumer.New(consumer, consumer.Topic),
	}
}
