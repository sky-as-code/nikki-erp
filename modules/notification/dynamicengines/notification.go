package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type notificationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_notification"`
}

// notificationLayerParam carries what the notification's own layers need beyond the composable
// defaults.
//
// The sibling repositories are injected rather than reached for, because the send writes a
// notification AND its recipients in one transaction: a resource that owns only its own table
// could not do that.
type notificationLayerParam struct {
	dig.In

	Recipients it.RecipientRepository
	Deliveries it.DeliveryRepository
	Broker     it.RealtimeNotificationBroker
	Logger     logging.LoggerService
}

func registerNotificationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.NotificationSchemaName),
		func(param composable.BuildParam, layers notificationLayerParam) composable.DynamicResourceEngineOnion {
			var notifications it.NotificationRepository

			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.NotificationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					notifications = repo.NewNotificationRepository(base)
					return notifications
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewNotificationDomainService(
						base, notifications, layers.Recipients, layers.Deliveries)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewNotificationApplicationService(base, layers.Broker, layers.Logger)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p notificationEngineParam) it.NotificationRepository {
			return p.Engine.Repository().(it.NotificationRepository)
		},
		func(p notificationEngineParam) it.NotificationDomainService {
			return p.Engine.DomainService().(it.NotificationDomainService)
		},
		func(p notificationEngineParam) it.NotificationApplicationService {
			return p.Engine.ApplicationService().(it.NotificationApplicationService)
		},
	))
}
