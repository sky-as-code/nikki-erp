package services

import (
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// NewVendorApplicationServiceImpl publishes the vendor capability under the name a CONSUMING
// module binds to.
//
// interfaces/vendor documents VendorAppService as "the type a consuming module's
// infra/external/index.go binds to its own local port", but nothing provided it — which is not a
// compile error, so it surfaced only when Purchase failed to start with "the vendor port is not
// registered". Registering the domain service alone left the documented contract unfulfilled.
//
// It lives here rather than in an app/ package because Contacts has none: the hand-written
// application layer was deleted when the module moved to the dynamic resource engine ([PUR-008]),
// and reviving a package for one delegation would be more structure than the delegation is worth.
// Essential, which kept its app/ layer, puts the equivalent there.
func NewVendorApplicationServiceImpl(
	vendorSvc itVendor.VendorDomainService,
) itVendor.VendorAppService {
	return &VendorApplicationServiceImpl{vendorSvc: vendorSvc}
}

// VendorApplicationServiceImpl stays a thin delegation on purpose: when Contacts is split into its
// own service, this is the type a REST client replaces, and any logic living here would have to be
// duplicated on the other side of the wire.
type VendorApplicationServiceImpl struct {
	vendorSvc itVendor.VendorDomainService
}

func (this *VendorApplicationServiceImpl) GetVendor(
	ctx corectx.Context, query itVendor.GetVendorQuery,
) (*itVendor.GetVendorResult, error) {
	return this.vendorSvc.GetVendor(ctx, query)
}

func (this *VendorApplicationServiceImpl) AssertOrderable(
	ctx corectx.Context, query itVendor.AssertOrderableQuery,
) (*itVendor.AssertOrderableResult, error) {
	return this.vendorSvc.AssertOrderable(ctx, query)
}
