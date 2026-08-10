package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func groupEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.GroupSchemaName,
		DefaultFields: []string{
			models.GroupFieldName,
			models.GroupFieldDescription,
		},
	}
}
