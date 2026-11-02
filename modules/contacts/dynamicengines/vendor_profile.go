package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type vendorProfileEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_contacts_vendor_profile"`
}

func registerVendorProfileEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.VendorProfileSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.VendorProfileSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewVendorProfileRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewVendorDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewVendorProfileApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p vendorProfileEngineParam) itVendor.VendorProfileRepository {
			return p.Engine.Repository().(itVendor.VendorProfileRepository)
		},
		func(p vendorProfileEngineParam) itVendor.VendorDomainService {
			return p.Engine.DomainService().(itVendor.VendorDomainService)
		},
		func(p vendorProfileEngineParam) itVendor.VendorProfileApplicationService {
			return p.Engine.ApplicationService().(itVendor.VendorProfileApplicationService)
		},
		// VendorAppService is the narrow capability Purchase binds to; it must keep resolving.
		func(p vendorProfileEngineParam) itVendor.VendorAppService {
			return p.Engine.ApplicationService().(itVendor.VendorAppService)
		},
	))
}
