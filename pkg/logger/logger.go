package logger

import (
	"log"

	"go.uber.org/zap"
)

var baseLog *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := cfg.Build()
	if err != nil {
		log.Fatal("failed to init log", err)
	}
	baseLog = logger
}

func NewLogger(name string) *zap.SugaredLogger {
	return baseLog.Sugar().Named(name)
}
