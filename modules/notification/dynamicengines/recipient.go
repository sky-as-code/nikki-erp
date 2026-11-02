package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type recipientEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_notification_recipient"`
	Config config.ConfigService
}

// registerRecipientEngine declares the recipient onion and publishes its typed layers.
//
// The inbox application service is built here rather than on the notification onion because
// everything it answers — what is unread, in what order, what has been read — is a property of the
// recipient row.
func registerRecipientEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.RecipientSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.RecipientSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewRecipientRepository(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p recipientEngineParam) it.RecipientRepository {
			return p.Engine.Repository().(it.RecipientRepository)
		},
		func(p recipientEngineParam) it.InboxApplicationService {
			return app.NewInboxApplicationService(
				p.Engine.ApplicationService(),
				p.Engine.Repository().(it.RecipientRepository),
				p.Config,
			)
		},
	))
}
