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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	AddProcess    string = "create"
	DeleteProcess        = "delete"
	OutboxProcess        = "outbox"
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

func (uc *UseCase) CreateUser(ctx context.Context, userID uuid.UUID) (err error) {
	// select
	err = uc.repo.SelectUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {

			// insert
			errCreate := uc.repo.CreateUser(ctx, userID)
			if errCreate != nil {

				if errors.Is(errCreate, repoerrs.ErrAlreadyExists) {
					return repoerrs.ErrAlreadyExists
				}
				if errors.Is(errCreate, repoerrs.ErrDB) {
					return usecase_errors.ErrDB
				}

				return errCreate
			} else {
				// user created
				return nil
			}

		}
		if errors.Is(err, repoerrs.ErrDB) {

			return usecase_errors.ErrDB
		}

		return err
	}

	// user exist in db
	return nil
}

func (uc *UseCase) AddOrDeleteUserSegment(ctx context.Context, input dto.AddToSegmentInput) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		isSlugsAddEmpty = len(input.SlugsAdd) == 0
		isSlugsDelEmpty = len(input.SlugsDel) == 0

		isSlugsAddLenGreter = len(input.SlugsAdd) > len(input.SlugsDel)

		conn *pgxpool.Conn
		tx   pgx.Tx

		operations []dto.Operation
	)

	conn, err = uc.repo.GetPool().Acquire(ctx)
	if err != nil {
		return err
	}
	defer func() {
		conn.Release()
	}()

	tx, err = conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted, // ? так быстрее, время важно, можно потом откатить при 2PC или SAGA
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return repoerrs.ErrDB
	}
	// realease conn or close tx.Conn().Close(ctx)
	defer func() {
		tx.Conn().Close(ctx)
	}()

	if isSlugsAddLenGreter {
		operations = make([]dto.Operation, 0, len(input.SlugsAdd))
	} else {
		operations = make([]dto.Operation, 0, len(input.SlugsDel))
	}

	if !isSlugsAddEmpty {
		ops, err := uc.processToSegments(ctx, tx, input, AddProcess)
		if err != nil {
			return err
		}
		operations = append(operations, ops...)
	}

	if !isSlugsDelEmpty {
		ops, err := uc.processToSegments(ctx, tx, input, DeleteProcess)
		if err != nil {
			return err
		}
		operations = append(operations, ops...)
	}

	err = uc.processOperationsOutbox(ctx, tx, operations)
	if err != nil {
		return err
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

func (uc *UseCase) processToSegments(ctx context.Context, tx pgx.Tx, input dto.AddToSegmentInput, process string) (operations []dto.Operation, err error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}

	var (
		segmentTxDTO = dto.SegmentTx{
			UserID:    input.UserID,
			CreatedAt: input.OperationAt,
			Operation: process,
			TTL:       input.TTL,
		}
	)

	if process == AddProcess {
		operations = make([]dto.Operation, 0, len(input.SlugsAdd))
		for _, slug := range input.SlugsAdd {
			segmentTxDTO.Slug = slug

			operation, err := uc.repo.SegmentTx(ctx, tx, segmentTxDTO)
			if err != nil {
				return nil, err
			}
			operations = append(operations, operation)
		}
	} else if process == DeleteProcess {
		operations = make([]dto.Operation, 0, len(input.SlugsDel))
		for _, slug := range input.SlugsDel {
			segmentTxDTO.Slug = slug

			operation, err := uc.repo.SegmentTx(ctx, tx, segmentTxDTO)
			if err != nil {
				return nil, err
			}
			operations = append(operations, operation)
		}
	}

	return operations, nil
}

func (uc *UseCase) processOperationsOutbox(ctx context.Context, tx pgx.Tx, operations []dto.Operation) (err error) {
	for _, op := range operations {
		operation := dto.SegmentTx{
			UserID:    op.UserID,
			Slug:      op.Segment,
			Operation: op.Operation,
			CreatedAt: op.OperationAt,
		}

		_, err := uc.repo.AddToOperationsOutboxTx(ctx, tx, operation)
		if err != nil {
			return err
		}
	}
	return nil
}
