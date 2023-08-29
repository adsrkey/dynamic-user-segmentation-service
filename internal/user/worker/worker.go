package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Worker struct {
	Pool  postgres.PgxPool
	Log   logger.Logger
	cache *Cache
}

type Cache struct {
	mu         *sync.RWMutex
	operations []dto.Operation
}

func New(pool postgres.PgxPool, log logger.Logger) *Worker {
	return &Worker{
		Pool: pool,
		Log:  log,
		cache: &Cache{
			operations: make([]dto.Operation, 0),
			mu:         &sync.RWMutex{},
		},
	}
}

func (cache *Cache) Add(operations []dto.Operation) {
	cache.mu.Lock()
	cache.operations = append(cache.operations, operations...)
	cache.mu.Unlock()
}

func (cache *Cache) DeleteFrom(indx int) {
	if indx != 0 && indx > 0 {
		cache.operations = cache.operations[indx:]
	}
}

func (cache *Cache) Clean() {
	cache.mu.Lock()
	cache.operations = make([]dto.Operation, 0)
	cache.mu.Unlock()
}

func (w *Worker) PingProcess(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("worker.PingProcess ctx done")
		case <-time.After(1 * time.Second):
			err := w.Pool.Ping(ctx)
			if err == nil {
				w.Log.Debug("worker.PingProcess w.Pool.Ping err nil")
				return nil
			}
			w.Log.Debug("worker.PingProcess w.Pool.Ping err !=nil")
		}
	}
}

func (w *Worker) ProcessFromCache(ctx context.Context) {
	w.PingProcess(ctx)

}

type addToOperationsOutboxTx func(ctx context.Context, tx pgx.Tx, operation dto.Operation) (operationID uuid.UUID, err error)

func (w *Worker) OperationProcess(
	ctx context.Context,
	slugs []string,
	operation string,
	operationsOutboxTxFunc addToOperationsOutboxTx,
	input dto.AddToSegmentInput,
	ErrAddCh chan struct{}) error {
	// если процесс не запишется в базу данных,
	// то хранить его где-то, а после переподключения перезаписать,
	// как вариант - передать как ивент в кафку и воркером оттуда брать, но пока так, через горутину:
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operations := make([]dto.Operation, 0, len(slugs))
	for _, slug := range slugs {
		// add successful transaction to operation outbox
		op := dto.Operation{
			UserID: input.UserID,
			// because slug == segment name
			Segment:     slug,
			Operation:   operation,
			OperationAt: time.Now(),
		}

		operations = append(operations, op)
	}

	outboxTx, err := w.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		w.Log.Error(err)
		w.cache.Add(operations)
		return err
	}

	for i := 0; i < len(operations); i++ {
		_, err = operationsOutboxTxFunc(ctx, outboxTx, operations[i])
		if err != nil {
			w.Log.Error(err)
			w.cache.Add(operations)
			return err
		}
	}

	// from cache
	w.cache.mu.Lock()
	if len(w.cache.operations) != 0 {
		for i := 0; i < len(w.cache.operations); i++ {
			_, err = operationsOutboxTxFunc(ctx, outboxTx, w.cache.operations[i])
			if err != nil {
				w.Log.Error(err)
				// помечаем operation для удаления
				w.cache.DeleteFrom(i)
				return err
			}
		}
		// clean
		w.cache.Clean()
	}
	w.cache.mu.Unlock()

	select {
	case _, ok := <-ErrAddCh:
		if ok {
			return fmt.Errorf("ErrAddCh closed")
		} else {
			w.Log.Debug("OK! Commit! add outbox")

			err = outboxTx.Commit(ctx)
			if err != nil {
				w.Log.Debug("outbox  outboxTx.Commit ", err)

				defer w.cache.Add(operations)
				errRollback := outboxTx.Rollback(ctx)
				if errRollback != nil {
					w.Log.Debug("outbox  outboxTx.Rollback ", err)
					return errRollback
				}
				return err
			}
			w.Log.Debug("outbox  outboxTx.Commit ", err)
		}
	case <-ctx.Done():
		w.Log.Debug("AddToSegments outbox ctx.Done")
		// TODO: то положим в кеш операцию, которая не записалась в базу данных
		w.cache.Add(operations)

		return fmt.Errorf("AddToSegments outbox ctx.Done")
	}

	return nil
}
