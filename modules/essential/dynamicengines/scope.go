package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// tenantScope is the permission scope of the tenant-wide reference catalogues (languages, enums,
// tags, model and module metadata): they carry no org_id, and the seeded resources say "tenant".
func tenantScope() *requestguard.ResourceScope {
	scope := requestguard.ResourceScopeTenant
	return &scope
}
