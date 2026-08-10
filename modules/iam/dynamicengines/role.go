package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func roleEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.RoleSchemaName,
		DefaultFields: []string{
			models.RoleFieldName,
			models.RoleFieldDescription,
		},
	}
}
