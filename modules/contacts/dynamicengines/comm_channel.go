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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type commChannelEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_contacts_comm_channel"`
}

func registerCommChannelEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.CommChannelSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.CommChannelSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewCommChannelRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewCommChannelDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
