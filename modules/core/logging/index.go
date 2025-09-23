package logging

import (
	"os"
	"strings"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

var logger LoggerService

func InitSubModule() {
	defaultLvl := LevelWarn
	logger = NewLogger(defaultLvl)
	levelEnv := strings.ToLower(os.Getenv(strings.ReplaceAll(string(c.LogLevel), ".", "_")))
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
