// Package party declares the party capability that Contacts offers to other modules.
//
// CRUD on the party itself goes through the dynamic resource engine and needs no types here. What
// lives in this package is the one question a module holding a party reference must be able to ask
// without reaching into Contacts' repositories: may this transaction name this party?
//
// A consuming module never imports this package from its domain or application layer. It declares a
// local port in its own interfaces/external/ and binds it once in infra/external/index.go — see
// docs/wiki/01 "Microservice-ready Monolith".
package party

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// AssertAssignableQuery asks whether a transaction in one organization may name this party.
//
// Both ids are required: a party belongs to exactly one organization, and the whole point of the
// check is to compare that against the organization the referring document lives in.
type AssertAssignableQuery struct {
	PartyId model.Id
	OrgId   model.Id

	// Field is the name the caller knows this reference by — "bill_to_party_id" on a sales order.
	// It is echoed back on any violation so the error points at the caller's own field rather than
	// at Contacts'.
	Field string
}

// GetFiscalIdentityQuery asks for the details that go on a tax document.
type GetFiscalIdentityQuery struct {
	PartyId model.Id
	OrgId   model.Id
}

// FiscalIdentity is who a party is for the purposes of a tax document.
//
// Deliberately not the whole party: a display name, an avatar or a job title have no place on an
// invoice, and a port that returned them would eventually put one there. Every field may be empty —
// a party can be recorded long before anybody knows its tax code — so a caller must decide whether
// what it got is enough for the document it is raising, rather than assuming.
type FiscalIdentity struct {
	TaxCode   string
	LegalName string
	Address   string
	Email     string
}
