package user

import (
	"context"
	"errors"
	"fmt"

	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) CreateUser(ctx context.Context, userID uuid.UUID) error {
	sql, args, _ := r.Builder.
		Insert("users").
		// Values(squirrel.Expr("DEFAULT")).
		Columns("id").
		Values(userID).
		Suffix("RETURNING id").
		ToSql()

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return repoerrs.ErrDB
	}

	var id uuid.UUID
	err = tx.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				return repoerrs.ErrAlreadyExists
			}
		}
		return fmt.Errorf("UserRepo.CreateUser - r.Pool.QueryRow: %v", err)
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
