package identity

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/identity/app"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/identity/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/identity/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &IdentityModule{}

type IdentityModule struct {
}

// LabelKey implements InCodeModule.
func (*IdentityModule) LabelKey() string {
	return "identity.moduleLabel"
}

// Name implements InCodeModule.
func (*IdentityModule) Name() string {
	return c.IdentityModuleName
}

// Deps implements InCodeModule.
func (*IdentityModule) Deps() []string {
	return []string{
		"settings",
	}
}

// Version implements InCodeModule.
func (*IdentityModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements InCodeModule.
func (*IdentityModule) Init() error {
	err := errors.Join(
		external.InitExternalServices(),
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)

	return err
}

// Init implements InCodeModule.
func (this *IdentityModule) RegisterModels() error {
	return errors.Join(
		// Identity
		dmodel.RegisterSchemaB(models.OrgUserRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.OrganizationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.OrganizationalUnitSchemaBuilder()),
		dmodel.RegisterSchemaB(models.GroupUserRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.GroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.UserSchemaBuilder()),

		// Authorize
		dmodel.RegisterSchemaB(models.ActionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ResourceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.EntitlementSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleGroupAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RoleUserAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.UserPermissionSchemaBuilder()),
	)
}
