package services

import (
	"fmt"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// tenantIdField is the tenant column the host's base model injects. IT Storage never declares it;
// it reads the acting tenant only to name the tenant's root space and object keys.
const tenantIdField = "tenant_id"

// actingTenantId is empty in a binary without tenants.
func actingTenantId(ctx corectx.Context) string {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil {
		return ""
	}
	switch raw := constraints[tenantIdField].(type) {
	case nil:
		return ""
	case string:
		return raw
	case *string:
		if raw == nil {
			return ""
		}
		return *raw
	default:
		return fmt.Sprint(raw)
	}
}
