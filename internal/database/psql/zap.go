package psql

import (
	"context"

	"github.com/AEKDA/ozon_task/internal/logger"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

type ZapLogger struct {
	Logger *logger.Logger
}

func (zl *ZapLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		zl.Logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		zl.Logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		zl.Logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		zl.Logger.Error(msg, fields...)
	default:
		zl.Logger.Info(msg, fields...)
	}
}
