package repository

import (
	"context"

	segmentDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"
)

type User interface {
	Pool

	CreateUser(ctx context.Context, userID uuid.UUID) error

	SelectUser(ctx context.Context, userID uuid.UUID) error
	SelectActiveUserSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)
	SelectSegmentID(ctx context.Context, slug string) (uuid.UUID, error)
	SelectReport(ctx context.Context, input userDTO.ReportInput) (reports []userDTO.Report, err error)

	SegmentTx
	AddUserSegmentToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation userDTO.SegmentTx) (operationID uuid.UUID, err error)

	TTL
}

type Segment interface {
	Pool

	CreateSegment(ctx context.Context, tx pgx.Tx, operation segmentDTO.Operation) (segmentID uuid.UUID, err error)

	DeleteSegment(ctx context.Context, operation userDTO.SegmentTx) (err error)

	TotalUserCount(ctx context.Context, operation segmentDTO.Operation) (result segmentDTO.Total, err error)

	SegmentTx
	AddSegmentToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation userDTO.SegmentTx) (operationID uuid.UUID, err error)
}

type Pool interface {
	GetPool() postgres.PgxPool
}

type SegmentTx interface {
	SegmentTx(ctx context.Context, tx pgx.Tx, input userDTO.SegmentTx) (operation userDTO.Operation, err error)
}

type TTL interface {
	SelectSegmentTTL(ctx context.Context, tx pgx.Tx, data userDTO.TTLTx) (results []userDTO.TTLTxR, err error)
	TTLMarkDone(ctx context.Context, tx pgx.Tx, data userDTO.TTLTx) (err error)
}
