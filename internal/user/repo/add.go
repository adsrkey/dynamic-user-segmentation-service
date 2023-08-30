package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) SegmentTx(ctx context.Context, tx pgx.Tx, input dto.SegmentTx) (operation dto.Operation, err error) {
	// Select segment_id
	selectedSegmentID, err := r.selectSegmentTx(ctx, tx, input.Slug)
	// rollback inside selectSegmentTx if err
	if err != nil {
		return dto.Operation{}, err
	}

	// TODO: select user_id if exists // create user
	var (
		segmentID uuid.UUID
		sql       string
		args      []any
		errMsg    string
	)

	if input.Operation == AddProcess {
		sql, args, err = r.Builder.
			Insert("segments_users").
			Columns("segment_id", "user_id").
			Values(selectedSegmentID, input.UserID).
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			return dto.Operation{}, err
		}

		errMsg = "err: add: "

	} else if input.Operation == DeleteProcess {
		sql, args, err = r.Builder.Delete("segments_users").
			Where(squirrel.Eq{"segment_id": selectedSegmentID, "user_id": input.UserID}). // TODO: обязателен индекс
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			return dto.Operation{}, err
		}

		errMsg = "err: delete: "
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&segmentID)
	if err != nil {

		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return dto.Operation{}, errRollback
		}

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrAlreadyExists.Error() + " user_id:" + input.UserID.String()
				return dto.Operation{}, errors.New(errMsg)
			}
			if pgErr.Code == "23503" {
				errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrNotFound.Error() + " user_id:" + input.UserID.String()
				return dto.Operation{}, errors.New(errMsg)
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			// TODO: logger
			r.Log.Debug(err)
			errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrNotFound.Error() + " user_id:" + input.UserID.String()
			return dto.Operation{}, errors.New(errMsg)
		}

		errMsg = "UserRepo.AddToOperationsOutboxTx - tx.QueryRow:" + err.Error()

		return dto.Operation{}, errors.New(errMsg)
	}

	op := dto.Operation{
		UserID:      input.UserID,
		Segment:     input.Slug,
		Operation:   input.Operation,
		OperationAt: input.CreatedAt,
	}

	return op, nil

}

func (r *Repo) AddToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation dto.SegmentTx) (operationID uuid.UUID, err error) {
	var (
		sql  string
		args []any
	)

	sql, args, err = r.Builder.
		Insert("operations_outbox").
		Columns("user_id", "segment", "operation", "operation_at").
		Values(operation.UserID, operation.Slug, operation.Operation, operation.CreatedAt).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.UUID{}, err
	}

	// Insert
	err = tx.QueryRow(ctx, sql, args...).Scan(&operationID)
	if err != nil {
		// r.Log.Debugf("err: %v", err)
		var errMsg string

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				errMsg = repoerrs.ErrAlreadyExists.Error() + " for the user with id:" + operation.UserID.String()
				return uuid.UUID{}, errors.New(errMsg)
			}
			if pgErr.Code == "23503" {
				errMsg = repoerrs.ErrNotFound.Error() + " user with id: " + operation.UserID.String()
				return uuid.UUID{}, errors.New(errMsg)
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return uuid.UUID{},
				fmt.Errorf("operation: '%s' with segment name: '%s' for the user with id: %s: %s to add",
					operation.Operation,
					operation.Slug,
					operation.UserID,
					repoerrs.ErrNotFound)
		}

		errMsg = "UserRepo.AddToOperationsOutboxTx - tx.QueryRow:" + err.Error()

		return uuid.UUID{}, errors.New(errMsg)
	}

	return operationID, nil
}
