package main

import (
	"github.com/famusovsky/rufkian-backend/internal/telephonist"
	"github.com/famusovsky/rufkian-backend/pkg/grace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// TODO update logger
	logCfg := zap.NewDevelopmentConfig()
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := logCfg.Build()
	defer logger.Sync()

	server := telephonist.NewServer(logger)

	grace.Handle(server, logger)
}
