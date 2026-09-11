package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// NewUomApplicationService is handed the composable default by the UoM onion.
func NewUomApplicationService(base composable.CrudApplicationService) itUom.UomApplicationService {
	uomSvc, ok := base.DomainService().(itUom.UomDomainService)
	if !ok {
		panic(errors.New("the uom onion must be built with NewUomDomainService"))
	}
	return &UomApplicationServiceImpl{CrudApplicationService: base, uomSvc: uomSvc}
}

// UomApplicationServiceImpl is the authorized CRUD of the resource. The domain service is kept
// under its own type so a custom action added later reaches the module's methods directly.
type UomApplicationServiceImpl struct {
	composable.CrudApplicationService
	uomSvc itUom.UomDomainService
}
