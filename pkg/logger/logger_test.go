package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestLoggerConfigurationSuccess(t *testing.T) {
	if baseLog == nil {
		t.Fatal("Expected baseLog to be initialized, but it was nil")
	}

	core := baseLog.Core()
	if core == nil {
		t.Fatal("Expected logger core to be initialized, but it was nil")
	}

	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	if core.Enabled(level.Level()) != true {
		t.Fatalf("Expected logger level to be %v, but got %v", zap.DebugLevel, level.Level())
	}
}

func TestLoggerConfigurationError(t *testing.T) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"/invalid/path"}

	_, err := cfg.Build()

	if err == nil {
		t.Fatal("Expected error while building logger configuration, but got nil")
	}
}

func TestNewLoggerReturnsNonNilInstance(t *testing.T) {
	log := NewLogger("test")
	if log == nil {
		t.Fatal("Expected non-nil *zap.SugaredLogger instance, got nil")
	}
}
