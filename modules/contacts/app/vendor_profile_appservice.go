package app

import (
	"go.bryk.io/pkg/errors"

	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewVendorProfileApplicationService is handed the composable default by the vendor profile
// onion; the domain service assertion is a wiring check.
func NewVendorProfileApplicationService(
	base composable.CrudApplicationService,
) itVendor.VendorProfileApplicationService {
	vendorSvc, ok := base.DomainService().(itVendor.VendorDomainService)
	if !ok {
		panic(errors.New("the vendor profile onion must be built with NewVendorDomainService"))
	}
	return &VendorProfileApplicationServiceImpl{CrudApplicationService: base, vendorSvc: vendorSvc}
}

// VendorProfileApplicationServiceImpl is the authorized CRUD plus the capability boundary
// Purchase binds to (VendorAppService). The cross-module methods assert no permission: the caller
// is another module acting on its own authority.
type VendorProfileApplicationServiceImpl struct {
	composable.CrudApplicationService
	vendorSvc itVendor.VendorDomainService
}

func (this *VendorProfileApplicationServiceImpl) GetVendor(
	ctx corectx.Context, query itVendor.GetVendorQuery,
) (*itVendor.GetVendorResult, error) {
	return this.vendorSvc.GetVendor(ctx, query)
}

func (this *VendorProfileApplicationServiceImpl) AssertOrderable(
	ctx corectx.Context, query itVendor.AssertOrderableQuery,
) (*itVendor.AssertOrderableResult, error) {
	return this.vendorSvc.AssertOrderable(ctx, query)
}
