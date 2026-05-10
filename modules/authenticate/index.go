package authenticate

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	models "github.com/sky-as-code/nikki-erp/modules/authenticate/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &AuthenticateModule{}

type AuthenticateModule struct {
}

// LabelKey implements NikkiModule.
func (*AuthenticateModule) LabelKey() string {
	return "authenticate.moduleLabel"
}

// Name implements NikkiModule.
func (*AuthenticateModule) Name() string {
	return modconstants.AuthenticateModuleName
}

// Deps implements NikkiModule.
func (*AuthenticateModule) Deps() []string {
	return []string{
		"identity",
	}
}

// IsInternal implements InCodeModule.
func (*AuthenticateModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*AuthenticateModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*AuthenticateModule) Init() error {
	err := stdErr.Join(
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)

	return err
}

// RegisterModels registers dynamic model schemas for this module.
func (*AuthenticateModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.LoginAttemptSchemaBuilder()),
		dmodel.RegisterSchemaB(models.MethodSettingSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PasswordStoreSchemaBuilder()),
	)
}
