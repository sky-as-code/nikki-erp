package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)


func entitlementEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.EntitlementSchemaName,
		DefaultFields: []string{
			models.EntitlementFieldName,
			models.EntitlementFieldDescription,
			models.EntitlementFieldExpression,
		},
	}
}
