package repository

import (
	"context"

	segmentRepo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo/segment"
	userRepo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/google/uuid"
)

type User interface {
	CreateUser(ctx context.Context, userID uuid.UUID) error
	SelectUser(ctx context.Context, userID uuid.UUID) error
	SelectActiveUserSegments(ctx context.Context, userID uuid.UUID) (slugs []string, err error)
	AddToSegments(ctx context.Context, slugsAdd []string, userID uuid.UUID) (slugs []string, err error)
	DeleteFromSegments(ctx context.Context, slugsDel []string, userID uuid.UUID) (err error)
	SelectSegmentID(ctx context.Context, slug string) (uuid.UUID, error)
}

type Segment interface {
	Create(ctx context.Context, slug string) (segmentID uuid.UUID, err error)
	Delete(ctx context.Context, slug string) (egmentID uuid.UUID, err error)
}

type Repositories struct {
	User
	Segment
}

func New(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		User:    userRepo.New(pg),
		Segment: segmentRepo.New(pg),
	}
}
