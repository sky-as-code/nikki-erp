package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/itstorage/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

type fileEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_itstorage_file"`
}

type fileEngineBuildParam struct {
	dig.In

	Build    composable.BuildParam
	SpaceSvc itSpace.SpaceDomainService
	Storage  filestorage.FileStorageAdapter
	QuotaSvc itExt.QuotaExtService
}

// registerFileEngine keeps the onion org-scoped: org_id is optional, but NULL means "the
// tenant's", which the built-in reads must not widen to every caller. The upload, meta and
// download methods authorize tenant-wide files themselves.
func registerFileEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.FileSchemaName),
		func(param fileEngineBuildParam) composable.DynamicResourceEngineOnion {
			domainDeps := services.FileDomainDeps{
				SpaceSvc:  param.SpaceSvc,
				Storage:   param.Storage,
				QuotaSvc:  param.QuotaSvc,
				ConfigSvc: param.Build.ConfigSvc,
			}
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.FileSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewFileRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewFileDomainService(base, domainDeps)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
