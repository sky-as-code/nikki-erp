package config

import (
	. "github.com/sky-as-code/nikki-erp/common/config/loader"
	"github.com/sky-as-code/nikki-erp/common/logging"
)

var configService *configServiceImpl

// Config loading priority, for example with key=DB_PASSWORD
// 1) Look for secret file path in key=DB_PASSWORD_FILE
//   - If there is a file mapped by K8s or Docker at this path, return the file content.
//   - Otherwise, proceed to 2).
//
// 2) Look for environment variable with name DB_PASSWORD, if not empty, return it.
// 3) Read in config/config.json, look for this key in current environment name group.
func InitSubModule() error {
	logger := logging.Logger()

	jsonLoader := NewJsonConfigLoader(logger)
	envVarLoader := NewEnvVarConfigLoader(jsonLoader, logger)
	secretFileLoader := NewSecretFileConfigLoader(envVarLoader, logger)
	configService = NewConfigService(secretFileLoader)

	err := configService.Init()
	return err
}

func ConfigSvcSingleton() ConfigService {
	return configService
}
