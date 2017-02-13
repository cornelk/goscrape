package appcontext

import (
	"time"

	"github.com/uber-go/zap"
)

var (
	// Logger is the global logger instance
	Logger zap.Logger
	// LogLevel is the log level that can be changed at runtime
	LogLevel zap.AtomicLevel
)

func init() {
	LogLevel = zap.DynamicLevel()
	LogLevel.SetLevel(zap.DebugLevel)
	Logger = zap.New(
		zap.NewTextEncoder(zap.TextTimeFormat(time.RFC3339)),
		zap.DebugLevel,
	)
}
