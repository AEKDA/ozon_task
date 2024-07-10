package app

import "github.com/AEKDA/ozon_task/internal/database/psql"

type Config struct {
	App struct {
		Host string `env:"APP_HOST" envDefault:""`
		Port uint32 `env:"APP_PORT" envDefault:"8080"`
	}
	Database    psql.Config
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	StorageType string `env:"STORAGE_TYPE" envDefault:"inmemory"`
}

const (
	TypeInmemory = "inmemory"
	TypePostgres = "postgres"
)
