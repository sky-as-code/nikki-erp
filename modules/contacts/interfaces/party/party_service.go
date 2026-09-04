package party

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AssertAssignableResult = dyn.OpResult[struct{}]
type GetFiscalIdentityResult = dyn.OpResult[FiscalIdentity]

// PartyLookupService reads a party for a purpose. Kept separate from the assertion so a caller that
// only validates a reference does not thereby gain the ability to read the party's tax details.
type PartyLookupService interface {
	// GetFiscalIdentity reads the details that go on a tax document raised against this party.
	//
	// Absence is reported as no data rather than as an error: the party may simply not exist in
	// this organization, which is an answer.
	GetFiscalIdentity(ctx corectx.Context, query GetFiscalIdentityQuery) (*GetFiscalIdentityResult, error)
}

// PartyDomainService is the full capability, implemented inside Contacts.
type PartyDomainService interface {
	PartyLookupService

	// AssertAssignable reports whether a document in this organization may name this party: the
	// party must exist, belong to that same organization, and not be archived.
	//
	// The archive rule applies to the WRITE only. A document that already names a party archived
	// afterwards keeps resolving it — history is not rewritten by a party going out of use, so
	// nothing re-validates a stored reference.
	AssertAssignable(ctx corectx.Context, query AssertAssignableQuery) (*AssertAssignableResult, error)
}

// PartyAppService is the capability other modules consume. It is the type a consuming module's
// infra/external/index.go binds to its own local port.
type PartyAppService interface {
	PartyLookupService

	AssertAssignable(ctx corectx.Context, query AssertAssignableQuery) (*AssertAssignableResult, error)
}
