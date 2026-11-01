package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// NewResourceCrudRepository is handed the composable default by the iam_resource onion.
func NewResourceCrudRepository(base dynamicresource.CrudRepository) itRes.ResourceCrudRepository {
	return &ResourceCrudRepositoryImpl{CrudRepository: base}
}

type ResourceCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}

// NewActionCrudRepository is handed the composable default by the iam_action onion.
func NewActionCrudRepository(base dynamicresource.CrudRepository) itAct.ActionCrudRepository {
	return &ActionCrudRepositoryImpl{CrudRepository: base}
}

type ActionCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}
