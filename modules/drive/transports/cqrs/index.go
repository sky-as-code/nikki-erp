package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initDriveFileHandler(),
	)
	return err
}

func initDriveFileHandler() error {
	deps.Register(NewDriveFileHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *DriveFileHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetDriveFileById),
		)
	})
}
