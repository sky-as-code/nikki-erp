package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
)

func InitExternalAdapter() error {
	err := stdErr.Join(
		deps.Register(file_storage.NewS3StorageService),
	)

	return err
}
