package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initUnitHandler(),
	)
	return err
}

func initUnitHandler() error {
	deps.Register(NewUnitHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *UnitHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateUnit),
			cqrs.NewHandler(handler.DeleteUnit),
			cqrs.NewHandler(handler.UpdateUnit),
			cqrs.NewHandler(handler.GetUnitById),
			cqrs.NewHandler(handler.SearchUnits),
		)
	})
}
