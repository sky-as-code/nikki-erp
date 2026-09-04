package services

import (
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// NewPartyApplicationServiceImpl publishes the party capability under the name a CONSUMING module
// binds to.
//
// Registering the domain service alone is not enough: interfaces/party documents PartyAppService as
// the type a consumer binds, and nothing providing it is not a compile error — it surfaces only at
// startup, as the consuming module failing to resolve its port. See the note on the vendor
// equivalent, where exactly that happened.
func NewPartyApplicationServiceImpl(
	partySvc itParty.PartyDomainService,
) itParty.PartyAppService {
	return &PartyApplicationServiceImpl{partySvc: partySvc}
}

// PartyApplicationServiceImpl stays a thin delegation on purpose: when Contacts is split into its
// own service, this is the type a REST client replaces, and any logic living here would have to be
// duplicated on the other side of the wire.
type PartyApplicationServiceImpl struct {
	partySvc itParty.PartyDomainService
}

func (this *PartyApplicationServiceImpl) AssertAssignable(
	ctx corectx.Context, query itParty.AssertAssignableQuery,
) (*itParty.AssertAssignableResult, error) {
	return this.partySvc.AssertAssignable(ctx, query)
}

func (this *PartyApplicationServiceImpl) GetFiscalIdentity(
	ctx corectx.Context, query itParty.GetFiscalIdentityQuery,
) (*itParty.GetFiscalIdentityResult, error) {
	return this.partySvc.GetFiscalIdentity(ctx, query)
}
