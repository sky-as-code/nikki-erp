package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type recipientEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_recipient"`
	Config config.ConfigService
}

// registerRecipientEngine declares the recipient onion and publishes its typed layers.
//
// The inbox application service is built here rather than on the notification onion because
// everything it answers — what is unread, in what order, what has been read — is a property of the
// recipient row.
func registerRecipientEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.RecipientSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.RecipientSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
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
