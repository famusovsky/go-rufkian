package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

type Driver string

type dbOptions struct {
	SSL bool
}

type Option func(*dbOptions)

func WithSSL(ssl bool) Option {
	return func(do *dbOptions) {
		do.SSL = ssl
	}
}

func Open(driver Driver, dsn string, options ...Option) (sqlx.Ext, error) {
	var cfg dbOptions
	for _, opt := range options {
		opt(&cfg)
	}

	dsn = fmt.Sprintf("%s://%s", driver, dsn)
	if !cfg.SSL {
		dsn = fmt.Sprintf("%s?sslmode=disable", dsn)
	}

	db, err := sql.Open(string(driver), dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, string(driver)), nil
}

func DsnFromEnv() string {
	dsn := fmt.Sprintf(
		"%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return dsn
}
