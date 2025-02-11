package main

import (
	"github.com/famusovsky/go-rufkian/internal/telephonist"
	"github.com/famusovsky/go-rufkian/pkg/grace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// TODO update logger
	logCfg := zap.NewDevelopmentConfig()
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := logCfg.Build()
	defer logger.Sync()

	// TODO use config
	server := telephonist.NewServer(logger, ":8080")

	grace.Handle(server, logger)
}
