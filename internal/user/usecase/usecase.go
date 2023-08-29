package user

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
	repo repo.User
	log  logger.Logger
}

func New(log logger.Logger, repo repo.User) *UseCase {
	return &UseCase{
		log:  log,
		repo: repo,
	}
}

func (uc *UseCase) AddToSegment(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error) {
	_, err = uc.repo.AddToSegments(ctx, input, process)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}
		return err
	}

	return nil
}

func (uc *UseCase) DeleteFromSegment(ctx context.Context, input dto.AddToSegmentInput, process *dto.Process) (err error) {
	err = uc.repo.DeleteFromSegments(ctx, input, process)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return usecase_errors.ErrDB
		}

		return err
	}

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

// TODO:
func (uc *UseCase) Reports(ctx context.Context, input dto.ReportInput) (reports []dto.Report, err error) {
	reports, err = uc.repo.SelectReport(ctx, input)
	if err != nil {
		if errors.Is(err, repoerrs.ErrDB) {
			return nil, usecase_errors.ErrDB
		}
		return nil, err
	}
	// TODO: запустить ftp или как там сервер файлов
	// TODO: создать csv файл и кинуть ссылку!

	// address := c.Echo().Server.Addr
	// link, err := linkgenerator.GenerateReportsLink(input)

	return reports, nil
}
