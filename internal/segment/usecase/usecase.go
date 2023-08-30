package segment

import (
	"context"
	"errors"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	usecase_errors "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/usecase/errors"
	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	repoerrs "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/errors"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/google/uuid"
)

type UseCase struct {
	log  logger.Logger
	repo repo.Segment
}

func New(log logger.Logger, repo repo.Segment) *UseCase {
	return &UseCase{
		log:  log,
		repo: repo,
	}
}

func (uc *UseCase) Create(ctx context.Context, operation dto.Operation) (segmentID uuid.UUID, err error) {
	segmentID, err = uc.repo.Create(ctx, operation)
	if err != nil {
		// TODO:
		if errors.Is(err, repoerrs.ErrDB) {
			return uuid.UUID{}, usecase_errors.ErrDB
		}
		return uuid.UUID{}, err
	}

	return segmentID, nil
}

func (uc *UseCase) Delete(ctx context.Context, operation dto.Operation) (err error) {
	err = uc.repo.Delete(ctx, operation)
	if err != nil {
		// TODO:
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}
		return err
	}
	// TODO: outbox

	return nil
}
