package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) DeleteFromSegments(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		cancel()
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	outboxTx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	for _, slug := range input.SlugsDel {
		err := r.deleteFromSegmentTx(ctx, tx, slug, input.UserID)
		if err != nil {
			process.ErrDelCh <- struct{}{}
			if errors.Is(err, repoerrs.ErrDB) {
				// TODO:
				// return usecase_errors.ErrDB
			}
			return err
		}

		go func(slug *string) {
			// add successful transaction to operation outbox
			op := dto.Operation{
				UserID: input.UserID,
				// because slug == segment name
				Segment:     *slug,
				Operation:   "delete",
				OperationAt: time.Now(),
			}
			_, err = r.addToOperationsOutboxTx(context.Background(), outboxTx, op)
			if err != nil {
				return
			}
			select {
			case _, ok := <-process.ErrAddCh:
				if ok {
					return
				} else {
					r.Log.Debug("OK! Commit! delete outbox")

					err = outboxTx.Commit(ctx)
					if err != nil {
						errRollback := outboxTx.Rollback(ctx)
						if errRollback != nil {
							return
						}
						return
					}
				}
			}
		}(&slug)
	}

	close(process.ErrDelCh)

	select {
	case _, ok := <-process.ErrAddCh:
		if ok {
			return fmt.Errorf("error")
		} else {
			r.Log.Debug("OK! Commit! add")
			err = tx.Commit(ctx)
			if err != nil {
				errRollback := tx.Rollback(ctx)
				if errRollback != nil {
					return errRollback
				}
				return err
			}
		}
	case <-ctx.Done():
		return fmt.Errorf("context done")
	}

	return nil
}

func (r *Repo) deleteFromSegmentTx(ctx context.Context, tx pgx.Tx, slugsDel string, userID uuid.UUID) (err error) {
	// Select

	// rollback inside select if err
	selectedSegmentID, err := r.selectSegmentTx(ctx, tx, slugsDel)
	if err != nil {
		return err
	}

	sql, args, err := r.Builder.Delete("segments_users").
		Where(squirrel.Eq{"segment_id": selectedSegmentID, "user_id": userID}). // TODO: обязателен индекс
		Suffix("RETURNING id").
		ToSql()

	var id uuid.UUID
	err = tx.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return errRollback
		}

		r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return repoerrs.ErrAlreadyExists
			}
			r.Log.Debugf(pgErr.Code)
		}
		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return fmt.Errorf("user with id: '%s' slug: '%s' %s to delete", userID, slugsDel, repoerrs.ErrNotFound)
		}

		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	if id == uuid.Nil {
		return fmt.Errorf("delete from segment uuid nil")
	}

	return nil
}
