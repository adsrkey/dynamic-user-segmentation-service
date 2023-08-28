package repository

import (
	"context"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"

	"github.com/google/uuid"
)

type User interface {
	SelectUser(ctx context.Context, userID uuid.UUID) error
	SelectActiveUserSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)
	AddToSegments(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (slugs []string, err error)
	DeleteFromSegments(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error)
	SelectSegmentID(ctx context.Context, slug string) (uuid.UUID, error)
}

type Segment interface {
	Create(ctx context.Context, slug string) (segmentID uuid.UUID, err error)
	Delete(ctx context.Context, slug string) (egmentID uuid.UUID, err error)
}
