package main

import (
	"flag"

	dbt "github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/pkg/database"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	dbWithSSL := flag.Bool("db_with_ssl", true, "connect to db with ssl")
	dotENV := flag.Bool("dot_env", false, "use env variables from .env file")
	flag.Parse()

	if *dotENV {
		godotenv.Load()
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := logCfg.Build()
	if err != nil {
		zap.L().Error("build logger config", zap.Error(err))
		return
	}
	defer logger.Sync()

	var db sqlx.Ext
	db, err = database.Open(database.PostgresDriver, database.DsnFromEnv(), database.WithSSL(*dbWithSSL))
	if err != nil {
		logger.Error("open database", zap.Error(err))
		return
	}

	dbClient := dbt.NewClient(db, logger)

	dbClient.UpdateUser(model.User{
		ID:        "5",
		Email:     "a@a.a",
		Password:  "123",
		Key:       "12345",
		TimeGoalM: 1,
	})
}
