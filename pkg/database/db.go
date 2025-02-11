package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

type Driver string

func Open(driver Driver, dsn string) (*sqlx.DB, error) {
	dsn = fmt.Sprintf("%s://%s", driver, dsn)

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
