package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/notification/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type deliveryEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_delivery"`
}

func registerDeliveryEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.DeliverySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.DeliverySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
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
