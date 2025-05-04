package core

import (
	"context"
	"errors"

	"github.com/sky-as-code/nikki-erp/common"
	"github.com/sky-as-code/nikki-erp/common/cqrs"
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/domain/user"
	"github.com/sky-as-code/nikki-erp/modules/core/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &SharedModule{}

type SharedModule struct {
}

// Name implements NikkiModule.
func (*SharedModule) Name() string {
	return "core"
}

// Deps implements NikkiModule.
func (*SharedModule) Deps() []string {
	return []string{common.ModuleSingleton.Name()}
}

// Init implements NikkiModule.
func (*SharedModule) Init() error {
	err := errors.Join(
		deps.Invoke(initUserSubModule),
		deps.Invoke(transport.InitTransport),
	)

	if err != nil {
		return err
	}
	return nil
}

func initUserSubModule() error {
	err := errors.Join(
		deps.Register(user.NewUserHandler),
		deps.Register(user.NewUserServiceImpl),
	)
	if err != nil {
		return err
	}
	return deps.Invoke(initUserHandlers)
}

func initUserHandlers(cqrsBus cqrs.CqrsBus, handler *user.UserHandler) error {
	ctx := context.Background()
	return cqrsBus.SubscribeRequests(
		ctx,
		cqrs.NewHandler(handler.Create),
	)
}
