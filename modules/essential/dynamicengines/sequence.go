package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// sequenceReadActions are the only built-in actions of counters and issuances: NextValue writes
// them with its own statements, and nothing else may.
func sequenceReadActions() []dynamicresource.CrudAction {
	return []dynamicresource.CrudAction{
		dynamicresource.CrudActionGetById, dynamicresource.CrudActionGetByUnique,
		dynamicresource.CrudActionSearch, dynamicresource.CrudActionExists, dynamicresource.CrudActionGetSchema,
	}
}

type sequenceDefinitionBuildParam struct {
	dynamicresource.BuildParam

	UsageDispatcher *usagecheck.Dispatcher
	SettingsSvc     itExt.EffectiveSettingsExtService
}

type sequenceEngineParam struct {
	dig.In

	Definition dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_definition"`
	Counter    dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_counter"`
	Issuance   dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_issuance"`
}

// registerSequenceEngines declares the three Sequence onions. A definition's org_id is optional (a
// tenant-wide definition has none), so org scoping is withdrawn and permissions are tenant-scoped,
// as for iam_role; counters and issuances inherit the definition's ownership and follow it.
func registerSequenceEngines() error {
	return stdErr.Join(
		registerSequenceDefinitionEngine(),
		registerSequenceCounterEngine(),
		registerSequenceIssuanceEngine(),
		registerSequenceTypedLayers(),
	)
}

func registerSequenceDefinitionEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SequenceDefinitionSchemaName),
		func(param sequenceDefinitionBuildParam) dynamicresource.DynamicResourceEngineOnion {
			ruleDeps := services.SequenceDefinitionDeps{
				UsageDispatcher: param.UsageDispatcher,
				DefaultTimezone: services.NewSequenceTimezoneDefaulter(param.SettingsSvc),
			}
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:             models.SequenceDefinitionSchemaName,
				IsOrgScoped:            util.ToPtr(false),
				PermissionScope:        tenantScope(),
				RejectArchivedOnCreate: true,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSequenceDefinitionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSequenceDefinitionDomainService(base, ruleDeps)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSequenceDefinitionApplicationService(base)
				},
			}, param.BuildParam)
		},
	)
}

func registerSequenceCounterEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SequenceCounterSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.SequenceCounterSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: tenantScope(),
				CrudActions:     sequenceReadActions(),
				IsAudited:       util.ToPtr(false),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSequenceCounterRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSequenceCounterDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSequenceCounterApplicationService(base)
				},
			}, param)
		},
	)
}

// The issuance is itself the record of what was issued, so it is not audited on top (IsAudited).
func registerSequenceIssuanceEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SequenceIssuanceSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.SequenceIssuanceSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: tenantScope(),
				CrudActions:     sequenceReadActions(),
				IsAudited:       util.ToPtr(false),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSequenceIssuanceRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSequenceIssuanceDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSequenceIssuanceApplicationService(base)
				},
			}, param)
		},
	)
}

func registerSequenceTypedLayers() error {
	return deps.Register(
		func(p sequenceEngineParam) itSeq.SequenceDefinitionRepository {
			return p.Definition.Repository().(itSeq.SequenceDefinitionRepository)
		},
		func(p sequenceEngineParam) itSeq.SequenceDefinitionDomainService {
			return p.Definition.DomainService().(itSeq.SequenceDefinitionDomainService)
		},
		func(p sequenceEngineParam) itSeq.SequenceDefinitionApplicationService {
			return p.Definition.ApplicationService().(itSeq.SequenceDefinitionApplicationService)
		},
		func(p sequenceEngineParam) itSeq.SequenceCounterRepository {
			return p.Counter.Repository().(itSeq.SequenceCounterRepository)
		},
		func(p sequenceEngineParam) itSeq.SequenceIssuanceRepository {
			return p.Issuance.Repository().(itSeq.SequenceIssuanceRepository)
		},
	)
}
