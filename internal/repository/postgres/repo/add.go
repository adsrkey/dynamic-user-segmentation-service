package repo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/Masterminds/squirrel"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) ttlTx(ctx context.Context, tx pgx.Tx, input userDTO.SegmentTx) (err error) {
	sql, args, err := r.Builder.
		Insert("ttl_segments").
		Columns("user_id", "segment_id", "ttl").
		Values(input.UserID, input.SegmentID, input.TTL).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		r.Log.Debug("Repo.ttlTx, r.Builder.Insert()", err)
		return err
	}

	var id uuid.UUID

	err = tx.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		r.Log.Debug("Repo.ttlTx, tx.QueryRow()", err)
		return err
	}
	return nil
}

func (r *Repo) SegmentTx(ctx context.Context, tx pgx.Tx, input userDTO.SegmentTx) (operation userDTO.Operation, err error) {
	var (
		selectedSegmentID uuid.UUID
	)

	if input.SegmentID != uuid.Nil {
		selectedSegmentID = input.SegmentID
	} else {
		selectedSegmentID, err = r.selectSegmentTx(ctx, tx, input.Slug)
		if err != nil {
			return userDTO.Operation{}, err
		}
	}

	var (
		segmentID uuid.UUID
		sql       string
		args      []any
		errMsg    string
	)

	if input.Operation == AddProcess {
		if !reflect.DeepEqual(input.TTL, time.Time{}) {
			input.SegmentID = selectedSegmentID
			err = r.ttlTx(ctx, tx, input)
			if err != nil {
				return userDTO.Operation{}, err
			}
		}

		sql, args, err = r.Builder.
			Insert("segments_users").
			Columns("segment_id", "user_id").
			Values(selectedSegmentID, input.UserID).
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			r.Log.Debug("Repo.SegmentTx, add_process, r.Builder.Insert()", err)
			return userDTO.Operation{}, err
		}

		errMsg = "err: add: "

	} else if input.Operation == DeleteProcess {
		sql, args, err = r.Builder.Delete("segments_users").
			Where(squirrel.Eq{"segment_id": selectedSegmentID, "user_id": input.UserID}). // TODO: обязателен индекс
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			r.Log.Debug("Repo.SegmentTx, delete_process, r.Builder.Insert()", err)
			return userDTO.Operation{}, err
		}

		errMsg = "err: delete: "
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&segmentID)
	if err != nil {
		r.Log.Debug("Repo.SegmentTx,  tx.QueryRow()", err)

		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			r.Log.Debug("Repo.SegmentTx, tx.Rollback()", err)
			return userDTO.Operation{}, errRollback
		}

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrAlreadyExists.Error() + " user_id:" + input.UserID.String()
				return userDTO.Operation{}, errors.New(errMsg)
			}
			if pgErr.Code == "23503" {
				errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrNotFound.Error() + " user_id:" + input.UserID.String()
				return userDTO.Operation{}, errors.New(errMsg)
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			errMsg = "slug with name:" + input.Slug + " " + errMsg + repoerrs.ErrNotFound.Error() + " user_id:" + input.UserID.String()
			return userDTO.Operation{}, errors.New(errMsg)
		}

		errMsg = "UserRepo.AddToOperationsOutboxTx - tx.QueryRow:" + err.Error()

		return userDTO.Operation{}, errors.New(errMsg)
	}

	op := userDTO.Operation{
		UserID:      input.UserID,
		Segment:     input.Slug,
		Operation:   input.Operation,
		OperationAt: input.CreatedAt,
	}

	return op, nil

}

func (r *Repo) AddUserSegmentToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation userDTO.SegmentTx) (operationID uuid.UUID, err error) {
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
		r.Log.Debug("Repo.AddUserSegmentToOperationsOutboxTx, r.Builder.Insert()", err)
		return uuid.UUID{}, err
	}

	// Insert
	err = tx.QueryRow(ctx, sql, args...).Scan(&operationID)
	if err != nil {
		r.Log.Debug("Repo.AddUserSegmentToOperationsOutboxTx, tx.QueryRow()", err)
		var errMsg string

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				errMsg = repoerrs.ErrAlreadyExists.Error() + "slug: " + operation.Slug + " for the user with id:" + operation.UserID.String()
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
