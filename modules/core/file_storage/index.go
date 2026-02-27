package file_storage

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitSubModule() error {
	err := stdErr.Join(
		deps.Register(NewS3StorageService),
	)

	return err
}
