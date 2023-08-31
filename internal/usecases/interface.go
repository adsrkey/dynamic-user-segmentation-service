package usecases

import (
	"context"

	segmentDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"

	"github.com/google/uuid"
)

type User interface {
	CreateUser(ctx context.Context, userID uuid.UUID) (err error)

	GetActiveSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)

	Reports(ctx context.Context, input userDTO.ReportInput) (report []userDTO.Report, err error)

	AddOrDeleteUserSegment(ctx context.Context, input userDTO.AddToSegmentInput) (err error)
}

type Segment interface {
	Create(ctx context.Context, operation segmentDTO.Operation) (segmentID uuid.UUID, err error)

	Delete(ctx context.Context, operation userDTO.SegmentTx) (err error)
}
