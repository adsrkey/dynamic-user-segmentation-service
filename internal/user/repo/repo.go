package user

import (
	worker "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/worker"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

type Repo struct {
	*postgres.Postgres
	worker *worker.Worker
}

func New(pg *postgres.Postgres) *Repo {
	return &Repo{pg, worker.New(pg.Pool, pg.Log)}
}
