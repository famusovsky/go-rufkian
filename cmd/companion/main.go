package main

import (
	"flag"

	"github.com/famusovsky/go-rufkian/internal/companion"
	"github.com/famusovsky/go-rufkian/pkg/database"
	"github.com/famusovsky/go-rufkian/pkg/grace"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	dbWithSSL := flag.Bool("db_with_ssl", true, "connect to db with ssl")
	dotENV := flag.Bool("dot_env", false, "use env variables from .env file")
	addr := flag.String("addr", ":8080", "http address")
	flag.Parse()

	if *dotENV {
		godotenv.Load()
	}

	// TODO update logger
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

	// TODO use config instead of addr
	server, err := companion.NewServer(logger, db, *addr)
	if err != nil {
		logger.Error("create server", zap.Error(err))
		return
	}

	grace.Handle(server, logger)
}
