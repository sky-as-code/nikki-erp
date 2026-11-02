package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

type languageEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_language"`
}

func registerLanguageEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.LanguageSchemaName),
		func(
			param dynamicresource.BuildParam, effectiveSettings itExt.EffectiveSettingsExtService,
			activation *services.LanguageActivation,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.LanguageSchemaName,
				PermissionScope: tenantScope(),
				IsOrgBound:      true,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewLanguageRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewLanguageDomainService(base, effectiveSettings, activation)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewLanguageApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p languageEngineParam) itLang.LanguageRepository {
			return p.Engine.Repository().(itLang.LanguageRepository)
		},
		func(p languageEngineParam) itLang.LanguageDomainService {
			return p.Engine.DomainService().(itLang.LanguageDomainService)
		},
		func(p languageEngineParam) itLang.LanguageApplicationService {
			return p.Engine.ApplicationService().(itLang.LanguageApplicationService)
		},
		// LanguageAppService is the narrow capability iam and module start-up bind to.
		func(p languageEngineParam) itLang.LanguageAppService {
			return p.Engine.ApplicationService().(itLang.LanguageAppService)
		},
	))
}
