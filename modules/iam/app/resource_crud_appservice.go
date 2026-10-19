package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// NewResourceCrudApplicationService is handed the composable default by the iam_resource onion.
func NewResourceCrudApplicationService(base composable.CrudApplicationService) itRes.ResourceCrudApplicationService {
	return &ResourceCrudApplicationServiceImpl{CrudApplicationService: base}
}

type ResourceCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// NewActionCrudApplicationService is handed the composable default by the iam_action onion.
func NewActionCrudApplicationService(base composable.CrudApplicationService) itAct.ActionCrudApplicationService {
	return &ActionCrudApplicationServiceImpl{CrudApplicationService: base}
}

type ActionCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
}
