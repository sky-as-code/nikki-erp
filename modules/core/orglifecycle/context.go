package orglifecycle

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ServiceContextOf builds the context a lifecycle command runs under.
//
// The bus carries no principal, so the handler acts as a SERVICE principal scoped to the org on
// the command. The inner context is kept: the bus puts the caller's domain constraints (the
// tenant, in the hosted app) on it, and a tenant-scoped repository panics without them.
func ServiceContextOf(ctx context.Context, moduleName string, orgId model.Id, requestedBy model.Id) corectx.Context {
	inner := ctx
	if reqCtx, ok := ctx.(corectx.Context); ok {
		inner = reqCtx.InnerContext()
	}

	org := orgId
	serviceCtx := corectx.NewRequestContextM(inner, moduleName)
	serviceCtx.SetPermissions(corectx.ContextPermissions{
		UserId: requestedBy,
		Principal: corectx.Principal{
			Kind:        corectx.PrincipalKindService,
			Id:          model.Id(moduleName),
			OrgId:       &org,
			DisplayName: moduleName + " org lifecycle",
		},
	})
	return serviceCtx
}
