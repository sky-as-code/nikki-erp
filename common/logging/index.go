package logging

import (
	"os"
	"strings"

	c "github.com/sky-as-code/nikki-erp/common/constants"
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
)

var logger LoggerService

func InitSubModule() {
	defaultLvl := LevelWarn
	logger = NewLogger(defaultLvl)
	levelEnv := strings.ToLower(os.Getenv(string(c.LogLevel)))
	logLevel, ok := levelNameMap[levelEnv]
	if !ok {
		logLevel = defaultLvl
		logger.Warnf("Value '%s' of the env var %s is not a valid log level. Fallback to level '%s'", levelEnv, c.LogLevel, defaultLvl)
	}
	logger.SetLevel(logLevel)
	err := deps.Register(func() LoggerService {
		return logger
	})
	if err != nil {
		panic(err)
	}
}

func Logger() LoggerService {
	if logger == nil {
		InitSubModule()
	}
	return logger
}
