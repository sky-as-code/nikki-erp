package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// PartyExtService answers two questions about a party: may this sale name it, and who is it for the
// purposes of a tax document.
//
// The second exists because a billing instruction holds only a REFERENCE to the party until the
// moment its invoice is raised. Nothing is copied while the arrangement is still being set up, so
// what the screen shows and what the document eventually says are the same thing read twice, and
// there is no copy sitting in between them going stale. The snapshot is taken once, at issuance,
// and the instruction is locked from that point on.
//
// Assignability is not derivable from any single field: it combines existence, organization and
// archival, and re-deriving it here would mean Sales keeps accepting a party Contacts has withdrawn
// the day a fourth condition is added.
type PartyExtService interface {
	// AssertAssignable answers whether a sale in the given organization may name this party.
	// Violations name the caller's own field, so an error points at `bill_to_party_id` rather than
	// at whatever Contacts calls its column.
	AssertAssignable(
		ctx corectx.Context, query AssertPartyAssignableQuery,
	) (*ft.ClientErrors, error)

	// GetFiscalIdentity reads who this party is for a tax document. Returns nil when there is no
	// such party in that organization, which a caller must treat as "cannot invoice" rather than
	// as an empty buyer.
	GetFiscalIdentity(
		ctx corectx.Context, query GetPartyFiscalIdentityQuery,
	) (*PartyFiscalIdentity, error)
}

// GetPartyFiscalIdentityQuery names the party whose details a document needs.
type GetPartyFiscalIdentityQuery struct {
	PartyId string
	OrgId   string
}

// PartyFiscalIdentity is who a party is on a tax document. Any field may be empty: a party can be
// recorded long before anybody knows its tax code.
type PartyFiscalIdentity struct {
	TaxCode   string
	LegalName string
	Address   string
	Email     string
}

// AssertPartyAssignableQuery names the party to check and the sale that would name it.
type AssertPartyAssignableQuery struct {
	PartyId string

	// OrgId is the ORDER's organization, not the party's. The check is precisely whether the two
	// agree, so passing the party's own would make it vacuously true.
	OrgId string

	// Field is the role being assigned — "sold_to_party_id", "bill_to_party_id", "payer_party_id"
	// — echoed back on a violation so the caller sees which role was refused.
	Field string
}
