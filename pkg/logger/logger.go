package logger

import (
	"log"

	"go.uber.org/zap"
)

var baseLog *zap.Logger

// init initializes the logger with the production configuration.
//
// It sets the logger's level to debug and disables stacktraces. If there is an error
// building the logger, it logs the error and exits the program.
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

// NewLogger creates a new SugaredLogger with the given name.
//
// Parameters:
// - name: the name of the logger.
//
// Returns:
// - *zap.SugaredLogger: a new SugaredLogger with the given name.
func NewLogger(name string) *zap.SugaredLogger {
	return baseLog.Sugar().Named(name)
}
