package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type partyEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_contacts_party"`
}

func registerPartyEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.PartySchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.PartySchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewPartyRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewPartyDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewPartyApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p partyEngineParam) itParty.PartyRepository {
			return p.Engine.Repository().(itParty.PartyRepository)
		},
		func(p partyEngineParam) itParty.PartyDomainService {
			return p.Engine.DomainService().(itParty.PartyDomainService)
		},
		func(p partyEngineParam) itParty.PartyApplicationService {
			return p.Engine.ApplicationService().(itParty.PartyApplicationService)
		},
		// PartyAppService is the narrow capability Sales binds to; it must keep resolving.
		func(p partyEngineParam) itParty.PartyAppService {
			return p.Engine.ApplicationService().(itParty.PartyAppService)
		},
	))
}
