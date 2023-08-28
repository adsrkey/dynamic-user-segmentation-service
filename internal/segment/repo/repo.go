package segment

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
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

func (r *Repo) Create(ctx context.Context, slug string) (segmentID uuid.UUID, err error) {
	sql, args, _ := r.Builder.
		Insert("segments").
		Columns("slug").
		Values(slug).
		Suffix("RETURNING id").
		ToSql()

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&segmentID)
	if err != nil {
		r.Log.Debugf("err: %v", err)

		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == "23505" {
				msg := repoerrs.ErrAlreadyExists.Error()
				return uuid.UUID{}, fmt.Errorf("slug with name: '%s' %s", slug, msg)
			}
		}
		return uuid.UUID{}, repoerrs.ErrDB
	}

	return segmentID, nil
}

func (r *Repo) Delete(ctx context.Context, slug string) (segmentID uuid.UUID, err error) {
	sql, args, _ := r.Builder.
		Delete("segments").
		Where(squirrel.Eq{"slug": slug}).
		Suffix("RETURNING id").
		ToSql()

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&segmentID)
	if err != nil {
		r.Log.Debugf("err: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			msg := repoerrs.ErrDoesNotExist.Error()
			return uuid.UUID{}, fmt.Errorf("%s slug with name: '%s' to delete", msg, slug)
		}
		return uuid.UUID{}, repoerrs.ErrDB
	}

	return segmentID, nil
}
