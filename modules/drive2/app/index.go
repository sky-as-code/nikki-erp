package app

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	return stdErr.Join(
		deps.Register(
			NewDriveFileService,
			NewDriveFileShareService,
			NewDriveFilePermissionService,
			NewDriveFileSignedUrlService,
		),
	)
}
