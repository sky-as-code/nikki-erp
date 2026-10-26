package external

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

type partyAdapter struct {
	parties itParty.PartyAppService
}

// AssertAssignable answers whether a sale may name this party, returning the violations alone.
//
// Contacts replies with an OpResult that carries its violations inside it; Sales wants them bare, so
// a party refusal joins the ones its own gates raise and the caller sees one list rather than having
// to merge two shapes.
func (this *partyAdapter) AssertAssignable(
	ctx corectx.Context, query itExt.AssertPartyAssignableQuery,
) (*ft.ClientErrors, error) {
	result, err := this.parties.AssertAssignable(ctx, itParty.AssertAssignableQuery{
		PartyId: model.Id(query.PartyId),
		OrgId:   model.Id(query.OrgId),
		Field:   query.Field,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "asserting party '%s' is assignable", query.PartyId)
	}
	if result == nil {
		// No answer is not an answer: treating it as a pass would let an unvalidated party onto the
		// order, which is the one outcome this port exists to prevent.
		return nil, errors.Errorf("asserting party '%s' is assignable returned nothing", query.PartyId)
	}
	if result.ClientErrors.Count() > 0 {
		return &result.ClientErrors, nil
	}
	return nil, nil
}

// GetFiscalIdentity reads who a party is for a tax document, or nil when there is no such party in
// that organization.
//
// Nil rather than a zero-valued identity, so a caller cannot mistake "no such party" for "a party
// with no tax code" — the first means the reference is wrong, the second that somebody must go and
// fill the details in.
func (this *partyAdapter) GetFiscalIdentity(
	ctx corectx.Context, query itExt.GetPartyFiscalIdentityQuery,
) (*itExt.PartyFiscalIdentity, error) {
	result, err := this.parties.GetFiscalIdentity(ctx, itParty.GetFiscalIdentityQuery{
		PartyId: model.Id(query.PartyId),
		OrgId:   model.Id(query.OrgId),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reading fiscal identity of party '%s'", query.PartyId)
	}
	if result == nil || !result.HasData {
		return nil, nil
	}

	return &itExt.PartyFiscalIdentity{
		TaxCode:   result.Data.TaxCode,
		LegalName: result.Data.LegalName,
		Address:   result.Data.Address,
		Email:     result.Data.Email,
	}, nil
}
