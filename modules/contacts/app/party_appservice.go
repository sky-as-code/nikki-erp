package app

import (
	"go.bryk.io/pkg/errors"

	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewPartyApplicationService is handed the composable default by the party onion. The domain
// service behind it is the module's own, so the assertion is a wiring check rather than a guess.
func NewPartyApplicationService(base composable.CrudApplicationService) itParty.PartyApplicationService {
	partySvc, ok := base.DomainService().(itParty.PartyDomainService)
	if !ok {
		panic(errors.New("the party onion must be built with NewPartyDomainService"))
	}
	return &PartyApplicationServiceImpl{CrudApplicationService: base, partySvc: partySvc}
}

// PartyApplicationServiceImpl is the authorized CRUD plus the capability boundary other modules
// bind to (PartyAppService).
//
// The cross-module methods stay a thin delegation on purpose: when Contacts is split into its own
// service, this is the type a REST client replaces. They assert no permission because the caller
// is another module acting on its own authority, not a user reading parties.
type PartyApplicationServiceImpl struct {
	composable.CrudApplicationService
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
