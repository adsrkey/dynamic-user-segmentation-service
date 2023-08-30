package ttl_worker

import (
	"context"
	"time"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

type TTLWorker struct {
	repo repo.User
}

func New(repo repo.User) *TTLWorker {
	return &TTLWorker{repo: repo}
}

func (worker *TTLWorker) DeleteUserFromSegment(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	now := time.Now().Format(time.RFC3339)

	ttlx := dto.TTLTx{
		TTL: now,
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

		_, err = worker.repo.SegmentTx(ctx, tx, input)
		if err != nil {
			return err
		}
	}

	err = worker.repo.DeleteSegmentTTL(ctx, tx, ttlx)
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
