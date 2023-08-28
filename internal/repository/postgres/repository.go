package repository

import (
	segmentRepo "github.com/adsrkey/dynamic-user-segmentation-service/internal/segment/repo"
	userRepo "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/repo"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

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
