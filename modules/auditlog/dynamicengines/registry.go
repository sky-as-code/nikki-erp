// Package dynamicengines declares the resource onion Audit Log serves through the composable
// resource engine, and registers it and its typed layers into the dependency container.
package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/app"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/auditlog/infra/repository"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func InitDynamicEngines() error {
	return registerActionEngine()
}

// ActionRestActions are the built-ins with a route. Entries are written by the AuditSink port
// only and never change, so clients only read.
func ActionRestActions() []dynamicresource.CrudAction {
	return []dynamicresource.CrudAction{
		dynamicresource.CrudActionGetById, dynamicresource.CrudActionSearch, dynamicresource.CrudActionExists,
		dynamicresource.CrudActionGetSchema,
	}
}

type actionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_auditlog_action"`
}

// registerActionEngine keeps create in the onion for the sink; update, delete and set_archived
// are left out, so an entry cannot be altered even by a direct service call. DeleteByOrg stays:
// it is the org cleanup, not a CRUD action. The resource is not audited, or every entry would
// write another.
func registerActionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ActionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.ActionSchemaName,
				CrudActions: append(ActionRestActions(), dynamicresource.CrudActionCreate),
				IsAudited:   util.ToPtr(false),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewActionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewActionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewActionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p actionEngineParam) itAction.ActionRepository {
			return p.Engine.Repository().(itAction.ActionRepository)
		},
		func(p actionEngineParam) itAction.ActionDomainService {
			return p.Engine.DomainService().(itAction.ActionDomainService)
		},
		func(p actionEngineParam) itAction.ActionApplicationService {
			return p.Engine.ApplicationService().(itAction.ActionApplicationService)
		},
	))
}
