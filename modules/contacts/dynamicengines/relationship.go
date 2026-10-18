package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type relationshipEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_relationship"`
}

func registerRelationshipEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.RelationshipSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.RelationshipSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewRelationshipRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewRelationshipDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewRelationshipApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p relationshipEngineParam) itRelationship.RelationshipRepository {
			return p.Engine.Repository().(itRelationship.RelationshipRepository)
		},
		func(p relationshipEngineParam) itRelationship.RelationshipDomainService {
			return p.Engine.DomainService().(itRelationship.RelationshipDomainService)
		},
		func(p relationshipEngineParam) itRelationship.RelationshipApplicationService {
			return p.Engine.ApplicationService().(itRelationship.RelationshipApplicationService)
		},
	))
}
