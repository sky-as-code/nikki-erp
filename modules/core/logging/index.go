package logging

import (
	"os"
)

var logLevelEnvName = "LIVE_LOG_LEVEL"
var logger LoggerService

func InitLogger() {
	logger = NewLogger(LVL_INFO)
	levelEnv := os.Getenv(logLevelEnvName)
	logLevel, ok := LvlMap[levelEnv]
	if !ok {
		logLevel = LVL_INFO
		logger.Warnf("Value '%s' of the env var %s is not a valid log level. Fallback to level '%s'", levelEnv, logLevelEnvName, LvlName[logLevel])
	}
	logger.SetLevel(logLevel)
}

func Logger() LoggerService {
	if logger == nil {
		InitLogger()
	}
	return logger
}
