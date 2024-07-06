package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	timeout  = time.Minute * 5
	maxConns = 20
)

func NewConnection(ctx context.Context, connCfg Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(connCfg.Parse())
	if err != nil {
		return nil, fmt.Errorf("config error %w", err)
	}
	cfg.MaxConns = maxConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init pool connect by config %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db %w", err)
	}

	return pool, nil
}
