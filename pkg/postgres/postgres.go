package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pgx also supported
)

type Config struct {
	Host     string
	Port     string
	User     string
	DBName   string
	SSLMode  string
	Password string
}

type Postgres struct {
	Sqlx *sqlx.DB
	// may add query builder
}

func NewPostgreSqlx(ctx context.Context, conf *Config) (*Postgres, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		conf.Host,
		conf.Port,
		conf.User,
		conf.DBName,
		conf.Password,
		conf.SSLMode,
	)
	DB, err := sqlx.ConnectContext(ctx, "postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		Sqlx: DB,
	}, nil
}
