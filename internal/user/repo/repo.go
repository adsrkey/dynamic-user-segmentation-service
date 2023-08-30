package user

import (
	worker "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/worker/operation"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

const (
	AddProcess    string = "create"
	DeleteProcess        = "delete"
	OutboxProcess        = "outbox"
)

type Repo struct {
	*postgres.Postgres
	worker *worker.OperationWorker
}

func New(pg *postgres.Postgres) *Repo {
	return &Repo{pg, worker.New(pg.Pool, pg.Log)}
}

func (r *Repo) Worker() *worker.OperationWorker {
	return r.worker
}
