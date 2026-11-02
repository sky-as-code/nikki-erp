package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// NewResourceCrudApplicationService is handed the composable default by the iam_resource onion.
func NewResourceCrudApplicationService(base dynamicresource.CrudApplicationService) itRes.ResourceCrudApplicationService {
	return &ResourceCrudApplicationServiceImpl{CrudApplicationService: base}
}

type ResourceCrudApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

// NewActionCrudApplicationService is handed the composable default by the iam_action onion.
func NewActionCrudApplicationService(base dynamicresource.CrudApplicationService) itAct.ActionCrudApplicationService {
	return &ActionCrudApplicationServiceImpl{CrudApplicationService: base}
}

type ActionCrudApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
