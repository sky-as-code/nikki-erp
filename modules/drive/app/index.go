package app

import "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitServices() error {
	return deps_inject.Register(
		NewDriveFileService,
		NewDriveFileShareService,
	)
}
