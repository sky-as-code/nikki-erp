package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func entitlementEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.EntitlementSchemaName,

		// org_id is conditional on an entitlement - required only when scope=org. A domain-scoped
		// entitlement legitimately has none, and hiding those would silently under-grant
		// permissions rather than report an error.
		WithdrawOrgScoping: true,
	}
}
