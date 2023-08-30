package usecases

import (
	"context"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"

	"github.com/google/uuid"
)

type User interface {
	CreateUser(ctx context.Context, userID uuid.UUID) (err error)

	Reports(ctx context.Context, input dto.ReportInput) (report []dto.Report, err error)
	GetActiveSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)

	// не доступна в интерфейсе
	// AddToSegment(ctx context.Context, input dto.AddToSegmentInput) (tx pgx.Tx, err error)

	AddOrDeleteUserSegment(ctx context.Context, input dto.AddToSegmentInput) (err error)

	// не доступна в интерфейсе
	// DeleteFromSegment(ctx context.Context, input dto.AddToSegmentInput) (err error)
}

type Segment interface {
	Create(ctx context.Context, operation dto.Operation) (segmentID uuid.UUID, err error)

	Delete(ctx context.Context, operation dto.Operation) (err error)
}
