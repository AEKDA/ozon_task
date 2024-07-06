package psql

import "fmt"

type Config struct {
	Name string `env:"DB_NAME" envDefault:"posts"`
	Port uint16 `env:"DB_PORT" envDefault:"5432"`
	User string `env:"DB_USER" envDefault:"postgres"`
	Pass string `env:"DB_PASS" envDefault:"postgres"`
	Host string `env:"DB_HOST" envDefault:"localhost"`
}

func (c *Config) Parse() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.User, c.Pass, c.Host, c.Port, c.Name)
}
