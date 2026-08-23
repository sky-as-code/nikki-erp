package iam

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/iam/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/iam/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/iam/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &IamModule{}

type IamModule struct {
}

// LabelKey implements InCodeModule.
func (*IamModule) LabelKey() string {
	return "iam.moduleLabel"
}

// Name implements InCodeModule.
func (*IamModule) Name() string {
	return c.IamModuleName
}

// Deps implements InCodeModule.
func (*IamModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"settings",
	}
}

// IsInternal implements InCodeModule.
func (*IamModule) IsInternal() bool {
	return false
}

// Version implements InCodeModule.
func (*IamModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements InCodeModule.
func (*IamModule) Init() error {
	// The resource engines must exist before the transport layer registers their routes.
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}

	err := errors.Join(
		external.InitExternalServices(),
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)

	return err
}

// OnAppStarted implements InCodeModuleAppStarted.
//
// The settings schema is registered here rather than in Init() because peer module init order is
// nondeterministic: Init() cannot assume the settings module has built its engines yet, while
// OnAppStarted runs after every module has initialized.
func (*IamModule) OnAppStarted() error {
	return deps.Invoke(func(settingsSvc itExt.SettingsRegistrationExtService) error {
		return registerTenantSettings(corectx.NewRequestContext(context.Background()), settingsSvc)
	})
}

// Init implements InCodeModule.
func (this *IamModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(models.OrgUserRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.OrganizationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.OrganizationalUnitSchemaBuilder()),
		dmodel.RegisterSchemaB(models.GroupUserRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.GroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.UserSchemaBuilder()),

		dmodel.RegisterSchemaB(models.ActionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ResourceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.EntitlementSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleGroupAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleUserAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.UserPermissionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PermissionHistorySchemaBuilder()),

		dmodel.RegisterSchemaB(models.LoginAttemptSchemaBuilder()),
		dmodel.RegisterSchemaB(models.MethodSettingSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PasswordStoreSchemaBuilder()),
	)
}
