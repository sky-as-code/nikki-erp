package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// grantRequestEngineSpec serves role requests, whose schema is named "iam_grant_request"
// and not "iam_role_request".
func grantRequestEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.RoleRequestSchemaName,
		DefaultFields: []string{
			models.RoleReqFieldStatus,
			models.RoleReqFieldType,
			models.RoleReqFieldRequestComment,
		},
	}
}
