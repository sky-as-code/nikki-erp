package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
)

type orgCurrencyEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_org_currency"`
}

type orgLanguageEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_org_language"`
}

// registerOrgActivationEngines declares the two activation junctions on the composable defaults.
// They serve no REST route: an activation is written through its currency's or language's own
// create and update, so these onions exist for their audited domain services alone.
func registerOrgActivationEngines() error {
	return stdErr.Join(
		registerPlainEngine(models.OrgCurrencySchemaName),
		registerPlainEngine(models.OrgLanguageSchemaName),
		deps.Register(
			func(p orgCurrencyEngineParam) *services.CurrencyActivation {
				return &services.CurrencyActivation{OrgActivation: services.NewOrgActivation(
					p.Engine.DomainService(), models.OrgCurrencyFieldCurrencyId)}
			},
			func(p orgLanguageEngineParam) *services.LanguageActivation {
				return &services.LanguageActivation{OrgActivation: services.NewOrgActivation(
					p.Engine.DomainService(), models.OrgLanguageFieldLanguageId)}
			},
		),
	)
}

func registerPlainEngine(schemaName string) error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(schemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{SchemaName: schemaName}, param)
		},
	)
}
