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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type relationshipEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_contacts_relationship"`
}

func registerRelationshipEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.RelationshipSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.RelationshipSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewRelationshipRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewRelationshipDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
