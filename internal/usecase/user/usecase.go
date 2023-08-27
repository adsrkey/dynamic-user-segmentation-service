package user

import (
	"context"
	"errors"

	"github.com/adsrkey/dynamic-user-segmentation-service/internal/domain"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/es/broker"
	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo/errors"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/google/uuid"
)

type UseCase struct {
	repo      repo.User
	log       logger.Logger
	msgBroker broker.User
}

func New(log logger.Logger, repo repo.User, msgBroker broker.User) *UseCase {
	return &UseCase{
		log:       log,
		repo:      repo,
		msgBroker: msgBroker,
	}
}

func (uc *UseCase) CreateUser(ctx context.Context, userID uuid.UUID) (err error) {
	// TODO: SELECT user
	err = uc.repo.SelectUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {

			errCreate := uc.repo.CreateUser(ctx, userID)
			if errCreate != nil {
				if errors.Is(errCreate, repoerrs.ErrAlreadyExists) {
					return repoerrs.ErrAlreadyExists
				}
				if errors.Is(errCreate, repoerrs.ErrDB) {
					return usecase_errors.ErrDB
				}
				return errCreate
			}
		}
		if err != nil {
			if errors.Is(err, repoerrs.ErrDB) {
				return usecase_errors.ErrDB
			}
		}

		return err
	}

	return nil
}

func (uc *UseCase) AddToSegment(ctx context.Context, slugsAdd []string, userID uuid.UUID) (err error) {
	// TODO: add kafka
	slugs, err := uc.repo.AddToSegments(ctx, slugsAdd, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}
		return err
	}

	go func(uc *UseCase) {
		uc.log.Debug("AddToSegment publish")
		err := uc.msgBroker.PublishUserAddToSegment(ctx, domain.User{
			ID:    userID,
			Slugs: slugs,
		})
		if err != nil {
			uc.log.Debug(err)
		}
	}(uc)

	return nil
}

func (uc *UseCase) DelFromSegment(ctx context.Context, slugsDel []string, userID uuid.UUID) (err error) {
	err = uc.repo.DeleteFromSegments(ctx, slugsDel, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}

		return err
	}

	go func(uc *UseCase) {
		uc.log.Debug("DelFromSegment publish")
		err := uc.msgBroker.PublishUserAddToSegment(ctx, domain.User{
			ID:    userID,
			Slugs: slugsDel,
		})
		if err != nil {
			uc.log.Debug(err)
		}
	}(uc)

	return nil
}

func (uc *UseCase) GetActiveSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error) {
	slugs, err = uc.repo.SelectActiveUserSegments(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return nil, usecase_errors.ErrDB
		}
		return nil, err
	}

	return slugs, nil
}
