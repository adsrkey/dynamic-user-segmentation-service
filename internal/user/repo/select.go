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
	"github.com/jackc/pgx/v5/pgxpool"
)

func (r *Repo) SelectUser(ctx context.Context, userID uuid.UUID) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		conn *pgxpool.Conn
		tx   pgx.Tx
	)

	conn, err = r.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer func() {
		conn.Release()
	}()

	tx, err = conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	err = r.selectUserTx(ctx, tx, userID)
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

func (r *Repo) selectUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	sql, args, err := r.Builder.Select("id").
		From("users").
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		return err
	}

	var selectedUserID uuid.UUID
	err = tx.QueryRow(ctx, sql, args...).Scan(&selectedUserID)
	if err != nil {
		r.Log.Debugf("err: %v", err)
		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return repoerrs.ErrNotFound
		}

		return repoerrs.ErrDB
	}

	return nil
}

func (r *Repo) SelectSegmentID(ctx context.Context, slug string) (id uuid.UUID, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		conn *pgxpool.Conn
		tx   pgx.Tx
	)

	conn, err = r.Pool.Acquire(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer conn.Release()

	tx, err = r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		r.Log.Error(err)
		return id, repoerrs.ErrDB
	}

	id, err = r.selectSegmentTx(ctx, tx, slug)
	if err != nil {
		return uuid.UUID{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return uuid.UUID{}, errRollback
		}
		return uuid.UUID{}, err
	}

	return id, nil
}

func (r *Repo) SelectActiveUserSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := r.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	sql, args, err := r.Builder.Select("slug").
		From("segments_users as su").
		Join("public.segments s ON s.id = su.segment_id").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()

	slugs = make([]string, 0, 0)

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// TODO:
	conn.Release()

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			return nil, err
		}
		slugs = append(slugs, n)
	}

	if err != nil {
		r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return nil, repoerrs.ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("UserRepo.CreateUser - r.Pool.QueryRow: %v", err)
	}

	return slugs, err
}

func (r *Repo) SelectReport(ctx context.Context, input dto.ReportInput) (reports []dto.Report, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := r.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		r.Log.Error(err)
		return nil, repoerrs.ErrDB
	}

	reports, err = r.selectReportTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return nil, errRollback
		}
		return nil, err
	}

	return reports, nil
}

func (r *Repo) SelectSegment(ctx context.Context, tx pgx.Tx, data dto.TTLTx) (results []dto.TTLTxR, err error) {
	sql, args, err := r.Builder.Select("user_id", "segment_id").
		From("ttl_segments").
		Where(squirrel.LtOrEq{"ttl": data.TTL}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	results = make([]dto.TTLTxR, 0)

	for rows.Next() {
		var result dto.TTLTxR
		if err := rows.Scan(&result.UserID, &result.SegmentID); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *Repo) selectSegmentTx(ctx context.Context, tx pgx.Tx, slug string) (uuid.UUID, error) {
	sql, args, err := r.Builder.Select("id").
		From("segments").
		Where(squirrel.Eq{"slug": slug}).
		ToSql()
	if err != nil {
		return uuid.UUID{}, err
	}

	var selectedSegmentID uuid.UUID
	err = tx.QueryRow(ctx, sql, args...).Scan(&selectedSegmentID)
	if err != nil {
		r.Log.Debugf("err: %v", err)

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return uuid.UUID{}, fmt.Errorf("segment with slug: %s %s", slug, repoerrs.ErrNotFound)
		}

		return uuid.UUID{}, repoerrs.ErrDB
	}

	return selectedSegmentID, nil
}

func (r *Repo) selectReportTx(ctx context.Context, tx pgx.Tx, input dto.ReportInput) (reports []dto.Report, err error) {

	operationAt := time.Date(input.Year, input.Month, 1, 0, 0, 0, 0, time.UTC)

	sql, args, err := r.Builder.Select("id", "user_id", "segment", "operation", "operation_at").
		From("operations_outbox").
		Where(squirrel.Eq{"user_id": input.UserID}, " AND operation_at > ", operationAt).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		r.Log.Debugf("err: %v", err)

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return nil, fmt.Errorf("report with user_id: %s %s", input.UserID, repoerrs.ErrNotFound)
		}

		return nil, repoerrs.ErrDB
	}

	reports = make([]dto.Report, 0, 1)

	for rows.Next() {
		var report dto.Report
		if err := rows.Scan(&report.ID, &report.UserID, &report.Segment,
			&report.Operation, &report.OperationAt); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (r *Repo) SelectSegmentTTL(ctx context.Context, tx pgx.Tx, data dto.TTLTx) (results []dto.TTLTxR, err error) {
	ttl, err := time.Parse(time.RFC3339, data.TTL)
	if err != nil {
		return nil, err
	}
	sql, args, err := r.Builder.Select("user_id", "segment_id").
		From("ttl_segments").
		Where(squirrel.LtOrEq{"ttl": ttl}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	results = make([]dto.TTLTxR, 0, 1)

	for rows.Next() {
		var result dto.TTLTxR
		if err := rows.Scan(&result.UserID, &result.SegmentID); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *Repo) DeleteSegmentTTL(ctx context.Context, tx pgx.Tx, data dto.TTLTx) (err error) {
	ttl, err := time.Parse(time.RFC3339, data.TTL)
	if err != nil {
		return err
	}

	sql, args, err := r.Builder.
		Delete("ttl_segments").
		Where(squirrel.Eq{"ttl": ttl}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
