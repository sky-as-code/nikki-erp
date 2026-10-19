package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type orgUnitEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_orgunit"`
}

// registerOrgUnitEngine declares the iam_orgunit onion: org-scoped, since a unit belongs to
// exactly one organization.
func registerOrgUnitEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.OrganizationalUnitSchemaName),
		func(param composable.BuildParam, userRepo itUser.UserCrudRepository) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.OrganizationalUnitSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewOrgUnitCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewOrgUnitCrudDomainService(base, userRepo)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					domSvc := base.DomainService().(itOrgUnit.OrgUnitCrudDomainService)
					return app.NewOrgUnitCrudApplicationService(base, domSvc)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p orgUnitEngineParam) itOrgUnit.OrgUnitCrudRepository {
			return p.Engine.Repository().(itOrgUnit.OrgUnitCrudRepository)
		},
		func(p orgUnitEngineParam) itOrgUnit.OrgUnitCrudDomainService {
			return p.Engine.DomainService().(itOrgUnit.OrgUnitCrudDomainService)
		},
		func(p orgUnitEngineParam) itOrgUnit.OrgUnitCrudApplicationService {
			return p.Engine.ApplicationService().(itOrgUnit.OrgUnitCrudApplicationService)
		},
	))
}
