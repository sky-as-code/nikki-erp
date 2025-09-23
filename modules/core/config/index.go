package config

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	. "github.com/sky-as-code/nikki-erp/modules/core/config/loader"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var configSvc ConfigService

// Config loading priority, for example with key=DB_PASSWORD
// 1) Look for secret file path in key=DB_PASSWORD_FILE
//   - If there is a file mapped by K8s or Docker at this path, return the file content.
//   - Otherwise, proceed to 2).
//
// 2) Look for environment variable with name DB_PASSWORD, if not empty, return it.
// 3) Read in config/config.json, look for this key in current environment name group.
func InitSubModule() (modErr error) {
	err := deps.Register(func(logger logging.LoggerService) ConfigService {
		var configService *configServiceImpl
		// jsonLoader := NewJsonConfigLoader(logger)
		yamlLoader := NewYamlConfigLoader(logger)
		envVarLoader := NewEnvVarConfigLoader(yamlLoader, logger)
		secretFileLoader := NewSecretFileConfigLoader(envVarLoader, logger)
		configService = NewConfigService(secretFileLoader)
		modErr = configService.Init()
		return configService
	})
	modErr = errors.Join(modErr, err)
	if modErr != nil {
		return
	}

	modErr = deps.Invoke(func(svc ConfigService) {
		configSvc = svc
	})

	return modErr
}

func ConfigSvcSingleton() ConfigService {
	return configSvc
}
