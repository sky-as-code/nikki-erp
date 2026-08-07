package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)


func orgEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.OrganizationSchemaName,
		DefaultFields: []string{
			models.OrgFieldDisplayName,
			models.OrgFieldLegalName,
			models.OrgFieldSlug,
			models.OrgFieldPhoneNumber,
		},
	}
}
