package ttl_worker

import (
	"context"
	"time"

	repo "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
)

const timeout = 30 * time.Second

type TTLWorker struct {
	repo repo.User
}

func New(repo repo.User) *TTLWorker {
	return &TTLWorker{repo: repo}
}

func (worker *TTLWorker) DeleteUserFromSegment(ctx context.Context) {
	for {
		time.Sleep(timeout)

		// TODO:

		// scan segment_user_ttl table
		// where ttl <= time.Now()

		// TODO: добавить индексы

		// worker.repo.DeleteFromSegments()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
