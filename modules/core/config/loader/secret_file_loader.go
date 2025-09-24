package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/env"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const SECRET_FILE_SUFFIX = "_FILE"

func NewSecretFileConfigLoader(nextLoader ConfigLoader, logger logging.LoggerService) *SecretFileConfigLoader {
	return &SecretFileConfigLoader{
		nextLoader,
		logger,
	}
}

// SecretFileConfigLoader gets configuration from a secret file,
// usually mapped by K8S of Docker to container.
type SecretFileConfigLoader struct {
	nextLoader ConfigLoader
	logger     logging.LoggerService
}

func (this *SecretFileConfigLoader) Init() error {
	appErr := this.nextLoader.Init()
	return appErr
}

func (fileLoader *SecretFileConfigLoader) Get(name string) (string, error) {
	hasSecretFile, secretFilePath := fileLoader.getSecretFilePath(name)
	if !hasSecretFile {
		return fileLoader.nextLoader.Get(name)
	}

	var fullPath string
	isRelativePath := strings.HasPrefix(secretFilePath, ".")
	if isRelativePath {
		workDir := env.Cwd()
		fullPath = filepath.Join(workDir, secretFilePath)
	} else {
		fullPath = secretFilePath
	}
	bytes, err := os.ReadFile(fullPath)

	if err != nil {
		return "", errors.Wrap(
			err,
			fmt.Sprintf("failed to read secret for key '%s' from file %s", name, secretFilePath),
		)
	}
	return string(bytes), nil
}

func (fileLoader *SecretFileConfigLoader) getSecretFilePath(name string) (bool, string) {
	// For example with name=CORE.DB.PASSWORD, we try getting the secret file path from config CORE.DB.PASSWORD_FILE.
	// If a value representing a path is specified, we read the secret content from that file path, otherwise we skip to next loader.

	fileNameConfig := fmt.Sprintf("%s%s", name, SECRET_FILE_SUFFIX)
	secretFilePath, err := fileLoader.nextLoader.Get(fileNameConfig)
	hasSecretFile := (err == nil)
	return hasSecretFile, secretFilePath
}
