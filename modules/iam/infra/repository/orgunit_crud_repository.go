package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

// NewOrgUnitCrudRepository is handed the composable default by the iam_orgunit onion.
func NewOrgUnitCrudRepository(base composable.CrudRepository) itOrgUnit.OrgUnitCrudRepository {
	return &OrgUnitCrudRepositoryImpl{CrudRepository: base}
}

type OrgUnitCrudRepositoryImpl struct {
	composable.CrudRepository
}
