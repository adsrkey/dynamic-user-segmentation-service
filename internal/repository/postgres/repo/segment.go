package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	dtoSegment "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/segment"
	userDTO "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) TotalUserCount(ctx context.Context, operation dtoSegment.Operation) (result dtoSegment.Total, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := r.Pool.Acquire(ctx)
	if err != nil {
		return dtoSegment.Total{}, err
	}
	defer func() {
		conn.Release()
	}()

	sql, args, err := r.Builder.
		Select("id").
		From("users").
		GroupBy("id").
		ToSql()
	if err != nil {
		return dtoSegment.Total{}, err
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})

	result.UserIDs = make([]uuid.UUID, 0)

	row, err := tx.Query(ctx, sql, args...)
	if err != nil {
		r.Log.Debugf("err: %v", err)
		return dtoSegment.Total{}, repoerrs.ErrDB
	}
	defer row.Close()

	for row.Next() {
		var userID uuid.UUID
		err := row.Scan(&userID)
		if err != nil {
			return dtoSegment.Total{}, err
		}
		result.UserIDs = append(result.UserIDs, userID)
	}

	result.TotalCount = len(result.UserIDs)

	if !tx.Conn().IsClosed() {
		err = tx.Commit(ctx)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return dtoSegment.Total{}, errRollback
			}
			return dtoSegment.Total{}, err
		}
	}

	return result, nil
}

func (r *Repo) CreateSegment(ctx context.Context, tx pgx.Tx, operation dtoSegment.Operation) (segmentID uuid.UUID, err error) {
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

	return segmentID, nil
}

func (r *Repo) AddSegmentToOperationsOutboxTx(ctx context.Context, tx pgx.Tx, operation userDTO.SegmentTx) (operationID uuid.UUID, err error) {
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

		errMsg = "SegmentRepo.AddToOperationsOutboxTx - tx.QueryRow:" + err.Error()

		return uuid.UUID{}, errors.New(errMsg)
	}

	return operationID, nil
}

func (r *Repo) DeleteSegment(ctx context.Context, operation userDTO.SegmentTx) (err error) {
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
		operations []userDTO.SegmentTx
	)

	operations = make([]userDTO.SegmentTx, 0)

	{
		sql, args, err := r.Builder.Select("s.id, su.user_id").
			From("segments as s").
			Join("public.segments_users su ON s.id = su.segment_id").
			Where(squirrel.Eq{"slug": operation.Slug}).
			ToSql()

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&segmentID, &operation.UserID); err != nil {
				return err
			}
			operations = append(operations, operation)
		}
		if len(operations) == 0 {
			return repoerrs.ErrDoesNotExist
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
		_, err := r.AddSegmentToOperationsOutboxTx(ctx, tx, operation)
		if err != nil {
			return err
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
