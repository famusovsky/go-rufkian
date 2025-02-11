package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func Open(driver, dsn string) (*sqlx.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, driver), nil
}

func DsnFromEnv(driver string) string {
	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s",
		driver,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return dsn
}
