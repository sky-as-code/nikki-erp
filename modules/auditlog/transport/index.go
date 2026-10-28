// Package transport subscribes Audit Log to the org lifecycle commands and mounts its routes.
package transport

import (
	"context"
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/transport/restful"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func InitTransport() error {
	return stdErr.Join(
		restful.InitRestfulHandlers(),
		initOrgLifecycleHandlers(),
	)
}

// initOrgLifecycleHandlers: a new org has no history to seed, and a removed org takes its entries
// with it. Global entries (org_id NULL) are untouched by an org's cleanup.
func initOrgLifecycleHandlers() error {
	svc := composable.NewOrgDataCleanup(models.ActionSchemaName)
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.AuditLogModuleName, svc)
	})
}
