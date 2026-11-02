package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type deliveryEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_notification_delivery"`
}

func registerDeliveryEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.DeliverySchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.DeliverySchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewDeliveryRepository(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p deliveryEngineParam) it.DeliveryRepository {
			return p.Engine.Repository().(it.DeliveryRepository)
		},
	))
}
