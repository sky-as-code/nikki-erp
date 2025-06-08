package loader

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/env"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var localEnvPath = filepath.Join("config", "local.env")

func NewEnvVarConfigLoader(jsonLoader *JsonConfigLoader, logger logging.LoggerService) *EnvVarConfigLoader {
	return &EnvVarConfigLoader{
		jsonLoader,
		logger,
	}
}

// EnvVarConfigLoader gets configuration from OS environment variables
type EnvVarConfigLoader struct {
	jsonLoader *JsonConfigLoader
	logger     logging.LoggerService
}

func (this *EnvVarConfigLoader) Init() (appErr error) {
	appErr = this.jsonLoader.Init()
	if appErr != nil {
		return
	}
	env.RunOnLocal(func() {
		this.logger.Infof("Local development detected! Loading local env file")
		if err := this.loadLocalEnvFile(); err != nil {
			appErr = err
		}
	})
	return
}

func (this *EnvVarConfigLoader) Get(name string) (string, error) {
	if val := os.Getenv(string(name)); len(val) > 0 {
		return val, nil
	}
	return this.jsonLoader.Get(name)
}

func (this *EnvVarConfigLoader) loadLocalEnvFile() error {
	workDir := env.Cwd()
	err := godotenv.Load(filepath.Join(workDir, localEnvPath))
	if err != nil && strings.Contains(err.Error(), "no such file") {
		this.logger.Warnf("Skipped loading env files because config/local.env doesn't exist. " +
			"Note that local.env file is excluded from Git commits, you must create it yourself " +
			"for local development to store passwords and other secrets")
		return nil
	}
	return errors.Wrap(err, "EnvVarConfigLoader.loadLocalEnvFile()")
}
