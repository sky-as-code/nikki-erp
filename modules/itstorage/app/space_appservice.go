package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

// NewSpaceApplicationService is handed the composable default by the itstorage_space onion. Spaces
// are system-managed, so only the authorized reads have routes.
func NewSpaceApplicationService(base composable.CrudApplicationService) itSpace.SpaceApplicationService {
	return &SpaceApplicationServiceImpl{CrudApplicationService: base}
}

type SpaceApplicationServiceImpl struct {
	composable.CrudApplicationService
}
