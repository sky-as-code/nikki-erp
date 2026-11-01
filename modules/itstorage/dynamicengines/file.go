package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/itstorage/infra/repository"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

type fileEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_itstorage_file"`
}

type fileEngineBuildParam struct {
	dig.In

	Build    dynamicresource.BuildParam
	SpaceSvc itSpace.SpaceDomainService
	Storage  filestorage.FileStorageAdapter
	QuotaSvc itExt.QuotaExtService
	// ConsumptionSvc is the typed layer, not the onion: the file onion only needs its counters.
	ConsumptionSvc itConsumption.ConsumptionDomainService
}

// registerFileEngine keeps the onion org-scoped: org_id is optional, but NULL means "the
// tenant's", which the built-in reads must not widen to every caller. The upload, meta and
// download methods authorize tenant-wide files themselves.
func registerFileEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.FileSchemaName),
		func(param fileEngineBuildParam) dynamicresource.DynamicResourceEngineOnion {
			domainDeps := services.FileDomainDeps{
				SpaceSvc:       param.SpaceSvc,
				Storage:        param.Storage,
				QuotaSvc:       param.QuotaSvc,
				ConsumptionSvc: param.ConsumptionSvc,
				ConfigSvc:      param.Build.ConfigSvc,
			}
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.FileSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewFileRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewFileDomainService(base, domainDeps)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewFileApplicationService(base)
				},
			}, param.Build)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p fileEngineParam) itFile.FileRepository {
			return p.Engine.Repository().(itFile.FileRepository)
		},
		func(p fileEngineParam) itFile.FileDomainService {
			return p.Engine.DomainService().(itFile.FileDomainService)
		},
		func(p fileEngineParam) itFile.FileApplicationService {
			return p.Engine.ApplicationService().(itFile.FileApplicationService)
		},
	))
}
