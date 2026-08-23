package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func orgUnitEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.OrganizationalUnitSchemaName,
	}
}
