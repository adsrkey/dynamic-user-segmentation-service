package usecases

import (
	"context"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"

	"github.com/google/uuid"
)

type User interface {
	AddToSegment(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error)
	DeleteFromSegment(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error)
	GetActiveSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)

	Reports(ctx context.Context, input dto.ReportInput) (report []dto.Report, err error)
}

type Segment interface {
	Create(ctx context.Context, slug string) (segmentID uuid.UUID, err error)
	Delete(ctx context.Context, slug string) (err error)
}
