package user

import (
	"context"
	"errors"
	"fmt"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) AddToSegments(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (slugs []string, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	slugs = make([]string, 0, 1)

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		r.Log.Error(err)
		return nil, repoerrs.ErrDB
	}

	for _, v := range input.SlugsAdd {
		_, err := r.addToSegmentTx(ctx, tx, v, input.UserID)
		if err != nil {
			if errors.Is(err, repoerrs.ErrDB) {
				// TODO:
				// return usecase_errors.ErrDB
			}
			process.ErrAddCh <- struct{}{}
			return nil, err
		}
		slugs = append(slugs, v)
	}

	select {
	case _, ok := <-process.ErrDelCh:
		if ok {
			return nil, fmt.Errorf("error")
		} else {
			r.Log.Debug("OK! Commit! add")
			err = tx.Commit(ctx)
			if err != nil {
				errRollback := tx.Rollback(ctx)
				if errRollback != nil {
					return nil, errRollback
				}
				return nil, err
			}
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("context done")
	}

	return slugs, nil
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
