package vendor

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type GetVendorResult = dyn.OpResult[GetVendorResultData]
type AssertOrderableResult = dyn.OpResult[struct{}]

// VendorLookupService answers questions about a single vendor. CRUD itself belongs to the dynamic
// resource engine; this exists because other modules must be able to validate a vendor reference
// without reaching into Contacts' repositories.
type VendorLookupService interface {
	GetVendor(ctx corectx.Context, query GetVendorQuery) (*GetVendorResult, error)
}

// VendorDomainService is the full capability, implemented inside Contacts.
type VendorDomainService interface {
	VendorLookupService

	// AssertOrderable reports whether a new order may name this party as its vendor: the party
	// must have a vendor profile in this organization, and that profile must be active and
	// unarchived.
	//
	// It is deliberately narrower than GetVendor. A suspended or blacklisted vendor must stay
	// readable, because orders already placed against it still name it — so a historical order
	// keeps resolving its vendor while a new one cannot select it.
	AssertOrderable(ctx corectx.Context, query AssertOrderableQuery) (*AssertOrderableResult, error)
}

// VendorAppService is the capability other modules consume. It is the type a consuming module's
// infra/external/index.go binds to its own local port.
type VendorAppService interface {
	VendorLookupService

	AssertOrderable(ctx corectx.Context, query AssertOrderableQuery) (*AssertOrderableResult, error)
}
