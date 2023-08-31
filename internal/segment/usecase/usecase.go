package segment

import (
	"context"
	"errors"
	"math/rand"

	segmentDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/usecase/errors"
	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UseCase struct {
	log  logger.Logger
	repo repo.Segment
}

func New(log logger.Logger, repo repo.Segment) *UseCase {
	return &UseCase{
		log:  log,
		repo: repo,
	}
}

func (uc *UseCase) Create(ctx context.Context, operation segmentDTO.Operation) (segmentID uuid.UUID, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := uc.repo.GetPool().Acquire(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() {
		conn.Release()
	}()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return uuid.UUID{}, repoerrs.ErrDB
	}

	segmentID, err = uc.repo.CreateSegment(ctx, tx, operation)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return uuid.UUID{}, usecase_errors.ErrDB
		}
		return uuid.UUID{}, err
	}

	operation.SegmentID = segmentID

	segmentTxs := make([]userDTO.SegmentTx, 0)

	if operation.Operation == segmentDTO.CreateAutoProcess {
		countOfUsers, err := uc.repo.TotalUserCount(ctx, operation)
		if err != nil {
			return uuid.UUID{}, err
		}

		countToAdd := int(float64(operation.Percent)/float64(100)*float64(countOfUsers.TotalCount) + 0.5)

		if countToAdd < 1 {
			return uuid.UUID{}, usecase_errors.ToFewUsers
		}

		if len(countOfUsers.UserIDs) > 2 {
			rand.Shuffle(
				len(countOfUsers.UserIDs),
				func(i, j int) {
					countOfUsers.UserIDs[i], countOfUsers.UserIDs[j] = countOfUsers.UserIDs[j], countOfUsers.UserIDs[i]
				},
			)
		}

		i := 0
		for _, userID := range countOfUsers.UserIDs {
			if i == countToAdd {
				break
			}

			input := userDTO.SegmentTx{
				UserID:    userID,
				Slug:      operation.Segment,
				Operation: dto.AddProcess,
				CreatedAt: operation.OperationAt,
				SegmentID: operation.SegmentID,
			}

			_, err = uc.repo.SegmentTx(ctx, tx, input)
			if err != nil {
				return uuid.UUID{}, err
			}

			segmentTxs = append(segmentTxs, input)
			i++
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return uuid.UUID{}, errRollback
		}
		return uuid.UUID{}, err
	}

	outboxTx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})

	if len(segmentTxs) > 0 {
		for _, input := range segmentTxs {
			_, err := uc.repo.AddSegmentToOperationsOutboxTx(ctx, outboxTx, input)
			if err != nil {
				return uuid.UUID{}, err
			}
		}
		err = outboxTx.Commit(ctx)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return uuid.UUID{}, errRollback
			}
			return uuid.UUID{}, err
		}
	}

	return segmentID, nil
}

func (uc *UseCase) Delete(ctx context.Context, operation userDTO.SegmentTx) (err error) {
	err = uc.repo.DeleteSegment(ctx, operation)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}
		return err
	}

	return nil
}
