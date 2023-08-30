package repository

import (
	"context"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"
)

type User interface {
	CreateUser(ctx context.Context, userID uuid.UUID) error

	SelectUser(ctx context.Context, userID uuid.UUID) error
	SelectActiveUserSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)
	SelectSegmentID(ctx context.Context, slug string) (uuid.UUID, error)
	SelectReport(ctx context.Context, input dto.ReportInput) (reports []dto.Report, err error)

	SegmentTx(ctx context.Context, tx pgx.Tx, input dto.SegmentTx) (operation dto.Operation, err error)

	GetPool() postgres.PgxPool

	AddToOutbox
}

type Segment interface {
	Create(ctx context.Context, operation dto.Operation) (segmentID uuid.UUID, err error)

	Delete(ctx context.Context, operation dto.Operation) (err error)
}

type AddToOutbox interface {
	AddToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation dto.SegmentTx) (operationID uuid.UUID, err error)
}
