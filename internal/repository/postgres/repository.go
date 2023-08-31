package repository

import (
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/repo"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

type Repositories struct {
	User
	Segment
}

func New(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		User:    repo.New(pg),
		Segment: repo.New(pg),
	}
}
