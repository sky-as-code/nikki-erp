package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// NewOrgCrudRepository is handed the composable default by the iam_org onion.
func NewOrgCrudRepository(base dynamicresource.CrudRepository) itOrg.OrgCrudRepository {
	return &OrgCrudRepositoryImpl{CrudRepository: base}
}

type OrgCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}
