package services

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitDomainServices() error {
	err := errors.Join(
		deps.Register(NewDriveFileAncestorDomainService),
		deps.Register(NewDriveFileDomainService),
		deps.Register(NewPermissionDomainService),
	)
	return err
}
