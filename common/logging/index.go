package logging

import (
	"os"

	c "github.com/sky-as-code/nikki-erp/common/constants"
)

var logger LoggerService

func InitSubModule() {
	defaultLvl := LevelWarn
	logger = NewLogger(defaultLvl)
	levelEnv := os.Getenv(string(c.LogLevel))
	logLevel, ok := levelNameMap[levelEnv]
	if !ok {
		logLevel = defaultLvl
		logger.Warnf("Value '%s' of the env var %s is not a valid log level. Fallback to level '%s'", levelEnv, c.LogLevel, defaultLvl)
	}
	logger.SetLevel(logLevel)
}

func Logger() LoggerService {
	if logger == nil {
		InitSubModule()
	}
	return logger
}
