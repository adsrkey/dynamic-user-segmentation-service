package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) AddToSegments(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (slugs []string, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r.Log.Debug("#1")

	slugs = make([]string, 0, len(input.SlugsAdd))

	r.Log.Debug("#2")

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return nil, repoerrs.ErrDB
	}

	r.Log.Debug("#3")

	for _, slug := range input.SlugsAdd {
		_, err := r.addToSegmentTx(ctx, tx, slug, input.UserID)
		if err != nil {
			if errors.Is(err, repoerrs.ErrDB) {
				// TODO:
				// return usecase_errors.ErrDB
			}
			process.ErrAddCh <- struct{}{}
			return nil, err
		}
		r.Log.Debug("#4")
		slugs = append(slugs, slug)
	}

	r.Log.Debug("#5")

	// но есть кейс когда в сегмент добавили, а в outbox нет
	// тогда можно вынести в worker, если соединение упало, то повторим транзакцию, когда соединение восстановится
	go func(slugs []string) {
		r.Log.Debug("#6")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		operation := "create"
		r.worker.OperationProcess(ctx, slugs, operation, r.addToOperationsOutboxTx, input, process.ErrAddCh)
		r.Log.Debug("#7")
	}(slugs)

	close(process.ErrAddCh)
	r.Log.Debug("#8")
	select {
	case _, ok := <-process.ErrDelCh:
		if ok {
			r.Log.Debug("#9")
			return nil, fmt.Errorf("error")
		} else {
			r.Log.Debug("#10")
			err = tx.Commit(ctx)
			if err != nil {
				r.Log.Debug("#11")
				r.Log.Debug("add tx.Commit", err)
				errRollback := tx.Rollback(ctx)
				if errRollback != nil {
					r.Log.Debug("#12")
					r.Log.Debug("add tx.Rollback", err)
					return nil, errRollback
				}
				r.Log.Debug("#13")
				return nil, err
			}
			r.Log.Debug("#14")

			r.Log.Debug("add tx.Committed")
		}
	case <-ctx.Done():
		r.Log.Debug("#15")
		return nil, fmt.Errorf("context done")
	}

	r.Log.Debug("#16")

	return slugs, nil
}

func (r *Repo) AddToOperationsOutbox(ctx context.Context, operation dto.Operation) (err error) {
	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	_, err = r.addToOperationsOutboxTx(ctx, tx, operation)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			// TODO:
			// return usecase_errors.ErrDB
		}
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

func (r *Repo) addToSegmentTx(ctx context.Context, tx pgx.Tx, slugsAdd string, userID uuid.UUID) (segmentID uuid.UUID, err error) {
	// Select segment_id
	selectedSegmentID, err := r.selectSegmentTx(ctx, tx, slugsAdd)
	// rollback inside selectSegmentTx if err
	if err != nil {
		return uuid.UUID{}, err
	}

	// TODO: select user_id if exists // create user

	sql2, args2, _ := r.Builder.
		Insert("segments_users").
		Columns("segment_id", "user_id").
		Values(selectedSegmentID, userID).
		Suffix("RETURNING id").
		ToSql()

	// Insert
	err = tx.QueryRow(ctx, sql2, args2...).Scan(&segmentID)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return uuid.UUID{}, errRollback
		}

		// r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				msg := repoerrs.ErrAlreadyExists.Error() + " for the user with id:" + userID.String()
				return uuid.UUID{}, fmt.Errorf("slug with name: '%s' %s", slugsAdd, msg)
			}
			if pgErr.Code == "23503" {
				msg := repoerrs.ErrNotFound.Error() + " user with id:"
				return uuid.UUID{}, fmt.Errorf("%s '%s'", msg, userID.String())
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return uuid.UUID{}, fmt.Errorf("user with id: %s slug: %s %s to add", userID, slugsAdd, repoerrs.ErrNotFound)
		}

		return uuid.UUID{}, fmt.Errorf("UserRepo.AddToSegment- r.Pool.QueryRow: %v", err)
	}

	// узнать с какими айдишниками эти названия slug
	return segmentID, nil

}

func (r *Repo) addToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation dto.Operation) (operationID uuid.UUID, err error) {
	sql2, args2, _ := r.Builder.
		Insert("operations_outbox").
		Columns("user_id", "segment", "operation", "operation_at").
		Values(operation.UserID, operation.Segment, operation.Operation, operation.OperationAt).
		Suffix("RETURNING id").
		ToSql()

		// Insert
	err = tx.QueryRow(ctx, sql2, args2...).Scan(&operationID)
	if err != nil {
		// r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				msg := repoerrs.ErrAlreadyExists.Error() + " for the user with id:" + operation.UserID.String()
				return uuid.UUID{}, fmt.Errorf("operation: '%s' with segment name: '%s': %s", operation.Operation, operation.Segment, msg)
			}
			if pgErr.Code == "23503" {
				msg := repoerrs.ErrNotFound.Error() + " user with id: " + operation.UserID.String()
				return uuid.UUID{}, fmt.Errorf("operation: '%s' with segment name: '%s': %s", operation.Operation, operation.Segment, msg)
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return uuid.UUID{},
				fmt.Errorf("operation: '%s' with segment name: '%s' for the user with id: %s: %s to add",
					operation.Operation,
					operation.Segment,
					operation.UserID,
					repoerrs.ErrNotFound)
		}

		return uuid.UUID{}, fmt.Errorf("UserRepo.addToOperationsOutboxTx- r.Pool.QueryRow: %v", err)
	}

	return operationID, nil
}
