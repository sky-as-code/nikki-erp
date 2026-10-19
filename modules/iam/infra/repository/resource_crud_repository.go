package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// NewResourceCrudRepository is handed the composable default by the iam_resource onion.
func NewResourceCrudRepository(base composable.CrudRepository) itRes.ResourceCrudRepository {
	return &ResourceCrudRepositoryImpl{CrudRepository: base}
}

type ResourceCrudRepositoryImpl struct {
	composable.CrudRepository
}

// NewActionCrudRepository is handed the composable default by the iam_action onion.
func NewActionCrudRepository(base composable.CrudRepository) itAct.ActionCrudRepository {
	return &ActionCrudRepositoryImpl{CrudRepository: base}
}

type ActionCrudRepositoryImpl struct {
	composable.CrudRepository
}
