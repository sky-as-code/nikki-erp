package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky-as-code/nikki-erp/common/logging"
	"github.com/sky-as-code/nikki-erp/common/util/env"
	. "github.com/sky-as-code/nikki-erp/common/util/fault"
)

const SECRET_FILE_SUFFIX = "_FILE"

func NewSecretFileConfigLoader(envVarLoader *EnvVarConfigLoader, logger logging.LoggerService) *SecretFileConfigLoader {
	return &SecretFileConfigLoader{
		envVarLoader,
		logger,
	}
}

// SecretFileConfigLoader gets configuration from a secret file,
// usually mapped by K8S of Docker to container.
type SecretFileConfigLoader struct {
	envVarLoader *EnvVarConfigLoader
	logger       logging.LoggerService
}

func (fileLoader *SecretFileConfigLoader) Init() AppError {
	appErr := fileLoader.envVarLoader.Init()
	return appErr
}

func (fileLoader *SecretFileConfigLoader) Get(name string) (string, AppError) {
	hasSecretFile, secretFilePath := fileLoader.getSecretFilePath(name)
	if !hasSecretFile {
		return fileLoader.envVarLoader.Get(name)
	}

	workDir := env.Cwd()
	bytes, err := os.ReadFile(filepath.Join(workDir, secretFilePath))

	if err != nil {
		return "", WrapTechnicalError(
			err,
			"SecretFileConfigLoader.Get('%s') failed to read secret from file %s", name, secretFilePath,
		)
	}
	return string(bytes), nil
}

func (fileLoader *SecretFileConfigLoader) getSecretFilePath(name string) (bool, string) {
	// For example with name=DB_PASSWORD, we try getting the secret file path from config DB_PASSWORD_FILE.
	// If the path is specified, we read the secret content from that file path, otherwise we
	// load DB_PASSWORD from env var.

	fileNameConfig := fmt.Sprintf("%s%s", string(name), SECRET_FILE_SUFFIX)
	secretFilePath, err := fileLoader.envVarLoader.Get(fileNameConfig)
	hasSecretFile := (err == nil)
	return hasSecretFile, secretFilePath
}
