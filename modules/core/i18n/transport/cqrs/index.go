package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initLanguageHandlers(),
	)
	return err
}

func initLanguageHandlers() error {
	deps.Register(NewLanguageHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *LanguageHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetCurrentLangCode),
			cqrs.NewHandler(handler.ListEnabledLangCodes),
		)
	})
}
