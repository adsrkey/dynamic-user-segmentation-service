package user

import (
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

type Repo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repo {
	return &Repo{pg}
}
