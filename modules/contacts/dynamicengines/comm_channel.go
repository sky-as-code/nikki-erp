package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type commChannelEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_comm_channel"`
}

func registerCommChannelEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.CommChannelSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.CommChannelSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewCommChannelRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewCommChannelDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewCommChannelApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p commChannelEngineParam) itCommChannel.CommChannelRepository {
			return p.Engine.Repository().(itCommChannel.CommChannelRepository)
		},
		func(p commChannelEngineParam) itCommChannel.CommChannelDomainService {
			return p.Engine.DomainService().(itCommChannel.CommChannelDomainService)
		},
		func(p commChannelEngineParam) itCommChannel.CommChannelApplicationService {
			return p.Engine.ApplicationService().(itCommChannel.CommChannelApplicationService)
		},
	))
}
