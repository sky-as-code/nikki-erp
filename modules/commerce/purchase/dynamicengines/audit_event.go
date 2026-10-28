package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/purchase/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The audit trail, written by the audit helper only; every client write is refused.

type auditEventEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_audit_event"`
}

func registerAuditEventEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.AuditEventSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.AuditEventSchemaName,
				// Not audited: purchase audit events are an audit trail of their own, written next to
				// the order change the audit log records.
				IsAudited: util.ToPtr(false),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewAuditEventRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewAuditEventCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewAuditEventApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p auditEventEngineParam) itAudit.AuditEventCrudRepository {
			return p.Engine.Repository().(itAudit.AuditEventCrudRepository)
		},
		func(p auditEventEngineParam) itAudit.AuditEventCrudDomainService {
			return p.Engine.DomainService().(itAudit.AuditEventCrudDomainService)
		},
		func(p auditEventEngineParam) itAudit.AuditEventCrudApplicationService {
			return p.Engine.ApplicationService().(itAudit.AuditEventCrudApplicationService)
		},
	))
}
