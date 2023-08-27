package segment

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
	log       logger.Logger
	repo      repo.Segment
	msgBroker broker.Segment
}

func New(log logger.Logger, repo repo.Segment, msgBroker broker.Segment) *UseCase {
	return &UseCase{
		log:       log,
		repo:      repo,
		msgBroker: msgBroker,
	}
}

func (uc *UseCase) Create(ctx context.Context, slug string) (segmentID uuid.UUID, err error) {
	segmentID, err = uc.repo.Create(ctx, slug)
	if err != nil {
		// TODO:
		if errors.Is(err, repoerrs.ErrDB) {
			return uuid.UUID{}, usecase_errors.ErrDB
		}
		return uuid.UUID{}, err
	}

	go func(uc *UseCase) {
		uc.log.Debug("Create segment publish")
		err := uc.msgBroker.PublishCreated(ctx, domain.Segment{
			ID:   segmentID,
			Slug: slug,
		})
		if err != nil {
			uc.log.Debug(err)
		}
	}(uc)

	return segmentID, nil
}

func (uc *UseCase) Delete(ctx context.Context, slug string) (err error) {
	segmentID, err := uc.repo.Delete(ctx, slug)
	if err != nil {
		// TODO:
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}
		return err
	}

	go func(uc *UseCase) {
		uc.log.Debug("Delete segment publish")
		err := uc.msgBroker.PublishDeleted(ctx, domain.Segment{
			ID:   segmentID,
			Slug: slug,
		})
		if err != nil {
			uc.log.Debug(err)
		}
	}(uc)
	return nil
}
