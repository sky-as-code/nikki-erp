package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

// NewOrgUnitCrudRepository is handed the composable default by the iam_orgunit onion.
func NewOrgUnitCrudRepository(base dynamicresource.CrudRepository) itOrgUnit.OrgUnitCrudRepository {
	return &OrgUnitCrudRepositoryImpl{CrudRepository: base}
}

type OrgUnitCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}
