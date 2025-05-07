package identity

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/app"
	repo "github.com/sky-as-code/nikki-erp/modules/identity/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/identity/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &IdentityModule{}

type IdentityModule struct {
}

// Name implements NikkiModule.
func (*IdentityModule) Name() string {
	return "identity"
}

// Deps implements NikkiModule.
func (*IdentityModule) Deps() []string {
	return []string{
		"core",
	}
}

// Init implements NikkiModule.
func (*IdentityModule) Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		initUserSubModule(),
		deps.Invoke(transport.InitTransport),
	)

	return err
}

func initUserSubModule() error {
	err := errors.Join(
		deps.Register(app.NewUserHandler),
		deps.Register(app.NewUserServiceImpl),
	)
	if err != nil {
		return err
	}
	return deps.Invoke(initUserHandlers)
}

func initUserHandlers(cqrsBus cqrs.CqrsBus, handler *app.UserHandler) error {
	ctx := context.Background()
	return cqrsBus.SubscribeRequests(
		ctx,
		cqrs.NewHandler(handler.Create),
	)
}
