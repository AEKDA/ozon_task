package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/AEKDA/ozon_task/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

const (
	timeout  = time.Minute * 5
	maxConns = 20
)

func NewConnection(ctx context.Context, connCfg Config, logger *logger.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(connCfg.Parse())
	if err != nil {
		return nil, fmt.Errorf("config error %w", err)
	}
	zapLogger := ZapLogger{logger}
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{Logger: &zapLogger, LogLevel: tracelog.LogLevelDebug}
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
