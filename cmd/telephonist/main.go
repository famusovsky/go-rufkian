package main

import (
	"flag"

	"github.com/famusovsky/go-rufkian/internal/telephonist"
	"github.com/famusovsky/go-rufkian/pkg/database"
	"github.com/famusovsky/go-rufkian/pkg/grace"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	withDB := flag.Bool("with_db", true, "store dialogs in db")
	flag.Parse()

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
	if *withDB {
		db, err = database.Open(database.PostgresDriver, database.DsnFromEnv())
		if err != nil {
			logger.Error("open database", zap.Error(err))
			return
		}
	} else {
		logger.Info("chosen mode without DB handling")
	}

	// TODO use config instead of addr
	server := telephonist.NewServer(logger, db, ":8080")

	grace.Handle(server, logger)
}
