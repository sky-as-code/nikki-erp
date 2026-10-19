package iam

import (
	"context"
	stdErr "errors"


	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
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

// ModelPrefix implements DynamicModule.
func (*IamModule) ModelPrefix() string {
	return c.IamModuleName
}

func (*IamModule) HasUi() bool {
	return true
}

func (*IamModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryFoundation
}

// Deps implements InCodeModule.
func (*IamModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"settings",
	}
}

// Dependants implements InCodeModuleDependants: the modules holding an org_id foreign key to
// iam_org, which IAM asks before deleting an organization. Hard-coded rather than derived, because
// each of these depends on iam, not the reverse.
//
// "parkinglot" ships only in the coreplatform binary. It is listed anyway, because the list
// belongs to IAM; the registry skips a dependant this binary did not load.
func (*IamModule) Dependants() []string {
	return []string{
		"parkinglot",
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

	err := stdErr.Join(
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
//
// The usage-check subscription is asserted here for the same reason: a dependant subscribes during
// its own Init, so this is the first point at which every one that was going to has done so.
func (*IamModule) OnAppStarted() error {
	return deps.Invoke(func(
		settingsSvc itExt.SettingsRegistrationExtService,
		cqrsBus cqrs.CqrsBus,
		dependants *modules.ModuleDependantRegistry,
	) error {
		if err := usagecheck.AssertDependantsSubscribed(cqrsBus, dependants, c.IamModuleName); err != nil {
			return err
		}
		return registerTenantSettings(corectx.NewRequestContext(context.Background()), settingsSvc)
	})
}

// Init implements InCodeModule.
func (this *IamModule) RegisterModels() error {
	return stdErr.Join(
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
