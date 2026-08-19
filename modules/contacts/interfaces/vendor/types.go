// Package vendor declares the vendor capability that Contacts offers to other modules.
//
// CRUD on the vendor profile itself goes through the dynamic resource engine and needs no types
// here. What lives in this package is the one question a purchasing module must be able to ask
// without reaching into Contacts' repositories: is this party a vendor I may order from, and what
// does it default?
//
// A consuming module never imports this package from its domain or application layer. It declares a
// local port in its own interfaces/external/ and binds it once in infra/external/index.go — see
// docs/wiki/01 "Microservice-ready Monolith".
package vendor

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// GetVendorQuery asks for the vendor profile of a party within an organization.
//
// Both ids are required because the profile is keyed by the pair: one party may be a vendor of one
// organization and an ordinary contact of another, with different terms in each.
type GetVendorQuery struct {
	PartyId model.Id
	OrgId   model.Id
}

// GetVendorResultData is what a purchasing module needs from a vendor, and nothing more.
//
// It deliberately does not expose the party — no display name, no tax id, no address. A caller that
// needs those is reading a contact, not validating a vendor, and should say so by reading the party
// resource through its own engine.
type GetVendorResultData struct {
	PartyId model.Id
	OrgId   model.Id
	Status  string

	// IsOrderable is the answer to the question the caller actually has. It is computed here
	// rather than left to the caller to derive from Status, so that the definition of "may be
	// ordered from" lives in one place: a caller comparing Status itself would have to be found
	// and changed when a fifth status is added.
	IsOrderable bool

	// DefaultCurrencyId, PaymentTerms and LeadTimeDays default the fields of a new order. All
	// three are optional — a vendor may be recorded before its terms are agreed — so a caller
	// must treat the zero value as "not stated" rather than as a value.
	DefaultCurrencyId *model.Id
	PaymentTerms      string
	LeadTimeDays      *int32
}

// AssertOrderableQuery asks whether a new order may name this party as its vendor.
type AssertOrderableQuery struct {
	PartyId model.Id
	OrgId   model.Id

	// Field is the name the caller knows this reference by — "vendor_id" on a purchase order. It
	// is echoed back on any violation so the error points at the caller's own field rather than
	// at Contacts'.
	Field string
}
