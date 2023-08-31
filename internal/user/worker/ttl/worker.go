package ttl_worker

import (
	"context"
	"time"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/jackc/pgx/v5"
)

type TTLWorker struct {
	repo repo.User
	log  logger.Logger
}

func New(repo repo.User, log logger.Logger) *TTLWorker {
	return &TTLWorker{repo: repo, log: log}
}

func (worker *TTLWorker) DeleteUserFromSegment(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		if err != nil {
			worker.log.Debug(err.Error())
		}
	}()

	conn, err := worker.repo.GetPool().Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}

	now := time.Now()

	nowFormat := now.Format(time.RFC3339)

	ttlx := dto.TTLTx{
		TTL: nowFormat,
	}

	results, err := worker.repo.SelectSegmentTTL(ctx, tx, ttlx)
	if err != nil {
		return err
	}

	for _, v := range results {
		input := dto.SegmentTx{}
		input.SegmentID = v.SegmentID
		input.Operation = dto.DeleteProcess
		input.UserID = v.UserID
		input.Slug = v.Slug

		op, err := worker.repo.SegmentTx(ctx, tx, input)
		if err != nil {
			return err
		}

		segmentTx := dto.SegmentTx{
			UserID:    input.UserID,
			Slug:      op.Segment,
			Operation: input.Operation,
			CreatedAt: now,
			TTL:       input.TTL,
			SegmentID: input.SegmentID,
		}

		_, err = worker.repo.AddUserSegmentToOperationsOutboxTx(ctx, tx, segmentTx)
		if err != nil {
			return err
		}
	}

	err = worker.repo.TTLMarkDone(ctx, tx, ttlx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return errRollback
		}
		return err
	}

	return nil
}
