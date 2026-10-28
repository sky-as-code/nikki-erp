package app

import (
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewActionApplicationService is handed the composable default by the auditlog_action onion. The
// built-in reads stay behind read:auditlog_action, for administrators; GetHistory is authorized
// by the target record instead.
func NewActionApplicationService(base composable.CrudApplicationService) itAction.ActionApplicationService {
	return &ActionApplicationServiceImpl{
		CrudApplicationService: base,
		access:                 resourceaccess.NewLazyResourceAccess(),
	}
}

type ActionApplicationServiceImpl struct {
	composable.CrudApplicationService

	// access is resolved lazily: the engine registers it in its own Init, which may run after
	// this onion is built.
	access resourceaccess.ResourceAccess
}
