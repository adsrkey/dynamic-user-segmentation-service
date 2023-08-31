package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/utils/repeatable"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

type PgxPool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Log     logger.Logger
	Builder squirrel.StatementBuilderType

	Pool PgxPool
}

func New(cfg config.PG, log logger.Logger) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
		Log:          log,
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("pgdb - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.PoolMax)

	err = repeatable.DoWithTries(func() error {
		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		ctx, cancel := context.WithTimeout(context.Background(), pg.connTimeout)
		defer cancel()

		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			log.Error(err)
			return err
		}

		err := pg.Pool.Ping(ctx)
		if err != nil {
			return err
		}

		return nil
	}, &pg.connAttempts, pg.connTimeout)

	if err != nil {
		return nil, fmt.Errorf("pgdb - New - pgxpool.ConnectConfig: %w", err)
	}

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
