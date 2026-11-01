package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type notificationEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_notification_notification"`
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
	Normalizer *services.SendNormalizer
	FanOut     *app.ChannelFanOut
	Logger     logging.LoggerService
}

func registerNotificationEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.NotificationSchemaName),
		func(param dynamicresource.BuildParam, layers notificationLayerParam) dynamicresource.DynamicResourceEngineOnion {
			var notifications it.NotificationRepository

			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.NotificationSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					notifications = repo.NewNotificationRepository(base)
					return notifications
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewNotificationDomainService(
						base, notifications, layers.Recipients, layers.Deliveries, layers.Normalizer)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewNotificationApplicationService(base, layers.Broker, layers.FanOut, layers.Logger)
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
