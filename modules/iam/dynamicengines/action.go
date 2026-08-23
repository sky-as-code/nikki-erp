package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func actionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ActionSchemaName,
	}
}
