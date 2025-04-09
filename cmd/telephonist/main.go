package main

import (
	"flag"
	"os"

	"github.com/famusovsky/go-rufkian/internal/telephonist"
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

	yaTranslateKey, yaFolderID := os.Getenv("YA_TRANSLATE_KEY"), os.Getenv("YA_FOLDER_ID")
	if yaTranslateKey == "" || yaFolderID == "" {
		logger.Error("yandex cloud credentials are not provided")
		return
	}

	companionURL := os.Getenv("COMPANION_URL")
	if companionURL == "" {
		logger.Error("companion url is not provided")
		return
	}

	// TODO use config instead of addr
	server := telephonist.NewServer(logger, db, *addr, companionURL, yaFolderID, yaTranslateKey)

	grace.Handle(server, logger)
}
