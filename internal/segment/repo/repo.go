package segment

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repo {
	return &Repo{pg}
}

func (r *Repo) Create(ctx context.Context, operation dto.Operation) (segmentID uuid.UUID, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := r.Pool.Acquire(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() {
		conn.Release()
	}()

	sql, args, err := r.Builder.
		Insert("segments").
		Columns("slug").
		Values(operation.Segment).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.UUID{}, err
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})

	err = tx.QueryRow(ctx, sql, args...).Scan(&segmentID)
	if err != nil {
		r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				msg := repoerrs.ErrAlreadyExists.Error()
				return uuid.UUID{}, fmt.Errorf("slug with name: '%s' %s", operation.Segment, msg)
			}
		}
		return uuid.UUID{}, repoerrs.ErrDB
	}

	if !tx.Conn().IsClosed() {
		err = tx.Commit(ctx)
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

func (r *Repo) AddToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation dto.Operation) (operationID uuid.UUID, err error) {
	var (
		sql  string
		args []any
	)

	sql, args, err = r.Builder.
		Insert("operations_outbox").
		Columns("user_id", "segment", "operation", "operation_at").
		Values(operation.UserID, operation.Segment, operation.Operation, operation.OperationAt).
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
					operation.Segment,
					operation.UserID,
					repoerrs.ErrNotFound)
		}

		errMsg = "SegmentRepo.AddToOperationsOutboxTx - tx.QueryRow:" + err.Error()

		return uuid.UUID{}, errors.New(errMsg)
	}

	return operationID, nil
}

func (r *Repo) Delete(ctx context.Context, operation dto.Operation) (err error) {
	conn, err := r.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer func() {
		conn.Release()
	}()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	var (
		segmentID  uuid.UUID
		operations []dto.Operation
	)

	operations = make([]dto.Operation, 0)

	{
		sql, args, err := r.Builder.Select("s.id, su.user_id").
			From("segments as s").
			Join("public.segments_users su ON s.id = su.segment_id").
			Where(squirrel.Eq{"slug": operation.Segment}).
			ToSql()

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return err
		}
		for rows.Next() {
			if err := rows.Scan(&segmentID, &operation.UserID); err != nil {
				return err
			}
			operations = append(operations, operation)
		}
	}

	{
		sql, args, err := r.Builder.
			Delete("segments").
			Where(squirrel.Eq{"id": segmentID}).
			ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}
	}

	for _, operation := range operations {
		_, err := r.AddToOperationsOutboxTx(ctx, tx, operation)
		if err != nil {
			return nil
		}
	}

	if !tx.Conn().IsClosed() {
		err = tx.Commit(ctx)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return errRollback
			}
			return err
		}
	}

	return nil
}
