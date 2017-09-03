package appcontext

import (
	"go.uber.org/zap"
)

var (
	// Logger is the global logger instance
	Logger zap.Logger
	// LogLevel is the log level that can be changed at runtime
	LogLevel zap.AtomicLevel
	// LogConf contains the global logger config
	LogConf zap.Config

	logger *zap.Logger
)

func init() {
	LogConf = zap.NewDevelopmentConfig()
	LogConf.Development = false
	LogConf.DisableCaller = true
	LogConf.DisableStacktrace = true
	LogLevel = LogConf.Level
	LogLevel.SetLevel(zap.InfoLevel)
	logger, _ = LogConf.Build()
}
