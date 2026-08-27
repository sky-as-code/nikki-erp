package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func roleEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.RoleSchemaName,

		// org_id is optional on a role: a role with no org accepts only domain-scoped entitlements
		// and is usable across organizations. Requiring org_id would make every such role
		// invisible, including during role assignment.
		WithdrawOrgScoping: true,
	}
}
