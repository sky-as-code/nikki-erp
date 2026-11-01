package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
)

// NewGroupCrudRepository is handed the composable default by the iam_group onion.
func NewGroupCrudRepository(base dynamicresource.CrudRepository) itGroup.GroupCrudRepository {
	return &GroupCrudRepositoryImpl{CrudRepository: base}
}

type GroupCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}
