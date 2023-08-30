package user

import (
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
)

func (r *Repo) GetPool() postgres.PgxPool {
	return r.Pool
}
