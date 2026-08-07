package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)


func resourceEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ResourceSchemaName,
		DefaultFields: []string{
			models.ResourceFieldName,
			models.ResourceFieldCode,
			models.ResourceFieldDescription,
		},
	}
}
