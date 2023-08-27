package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) SelectUser(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
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

func (r *Repo) SelectSegmentID(ctx context.Context, slug string) (uuid.UUID, error) {
	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return uuid.UUID{}, repoerrs.ErrDB
	}

	id, err := r.selectSegmentTx(ctx, tx, slug)
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
	sql, args, err := r.Builder.Select("slug").
		From("segments_users as su").
		Join("public.segments s ON s.id = su.segment_id").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()

	slugs = make([]string, 0, 0)

	rows, err := r.Pool.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}
	// TODO:
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

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return uuid.UUID{}, repoerrs.ErrAlreadyExists
			}
		}

		if ok := errors.Is(err, pgx.ErrNoRows); ok {
			r.Log.Debugf(pgx.ErrNoRows.Error())
			return uuid.UUID{}, fmt.Errorf("segment with slug: %s %s", slug, repoerrs.ErrNotFound)
		}

		return uuid.UUID{}, repoerrs.ErrDB
	}

	return selectedSegmentID, nil
}
