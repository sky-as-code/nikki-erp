package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// NewUomCatApplicationService is handed the composable default by the UoM Category onion.
func NewUomCatApplicationService(base dynamicresource.CrudApplicationService) itUomCat.UomCatApplicationService {
	uomCatSvc, ok := base.DomainService().(itUomCat.UomCatDomainService)
	if !ok {
		panic(errors.New("the uomcat onion must be built with NewUomCatDomainService"))
	}
	return &UomCatApplicationServiceImpl{CrudApplicationService: base, uomCatSvc: uomCatSvc}
}

type UomCatApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	uomCatSvc itUomCat.UomCatDomainService
}
