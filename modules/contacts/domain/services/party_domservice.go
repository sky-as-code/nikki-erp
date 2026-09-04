package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewPartyDomainServiceImpl() itParty.PartyDomainService {
	return &PartyDomainServiceImpl{}
}

type PartyDomainServiceImpl struct {
}

// AssertAssignable reports whether a document in one organization may name this party.
//
// The three failure cases are separate violations because they are different situations for whoever
// reads the error: a party that does not exist at all, one that exists but belongs to somebody
// else's organization, and one that is deliberately out of use.
//
// Cross-org is reported as "not found" rather than as its own reason. A caller in organization A
// must not be able to discover that a party id is real in organization B by reading the error, and
// from A's point of view a party it may not reference is indistinguishable from one that is absent.
func (this *PartyDomainServiceImpl) AssertAssignable(
	ctx corectx.Context, query itParty.AssertAssignableQuery,
) (*itParty.AssertAssignableResult, error) {
	vErrs := &ft.ClientErrors{}
	field := query.Field
	if field == "" {
		field = models.PartyFieldId
	}

	found, err := this.loadParty(ctx, query.PartyId, query.OrgId)
	if err != nil {
		return nil, errors.Wrap(err, "assert party assignable")
	}

	switch {
	case found == nil:
		vErrs.Append(*ft.NewBusinessViolation(field, "party.not_found",
			"no party exists with that id in this organization"))
	case isArchived(found):
		vErrs.Append(*ft.NewBusinessViolation(field, "party.archived",
			"the party is archived and cannot be named on a new transaction"))
	}

	if vErrs.Count() > 0 {
		return &itParty.AssertAssignableResult{ClientErrors: *vErrs}, nil
	}
	return &itParty.AssertAssignableResult{HasData: true}, nil
}

// GetFiscalIdentity reads the details that go on a tax document raised against this party.
//
// Read at the moment a document is raised rather than copied when the arrangement was made: what a
// tax authority must see is who the buyer is now, and a copy taken weeks earlier says who they were.
//
// Email is not returned. It lives on the comm_channel resource rather than the party, needs a typed
// lookup to pick the invoicing address out of several, and is optional on a VAT invoice — so it is
// left to a caller that actually needs it rather than half-guessed here.
func (this *PartyDomainServiceImpl) GetFiscalIdentity(
	ctx corectx.Context, query itParty.GetFiscalIdentityQuery,
) (*itParty.GetFiscalIdentityResult, error) {
	found, err := this.loadParty(ctx, query.PartyId, query.OrgId)
	if err != nil {
		return nil, errors.Wrap(err, "get party fiscal identity")
	}
	if found == nil {
		return &itParty.GetFiscalIdentityResult{HasData: false}, nil
	}

	party := models.NewPartyFrom(found)
	return &itParty.GetFiscalIdentityResult{
		Data: itParty.FiscalIdentity{
			TaxCode:   derefString(party.GetTaxId()),
			LegalName: derefString(party.GetLegalName()),
			Address:   derefString(party.GetLegalAddress()),
		},
		HasData: true,
	}, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// loadParty fetches a party constrained to one organization, returning nil when there is none.
//
// The org id is part of the FILTER rather than compared afterwards, so a party belonging to another
// organization is simply not found — the read never returns a row the caller may not see.
//
// A missing engine is a Go error rather than a validation failure: it means this module was
// initialized wrongly, which is a defect in the deployment and not something a caller can correct.
func (this *PartyDomainServiceImpl) loadParty(
	ctx corectx.Context, partyId model.Id, orgId model.Id,
) (dmodel.DynamicFields, error) {
	engine, ok := dynamicresource.Registry().GetEngine(models.PartySchemaName)
	if !ok {
		return nil, errors.Errorf("loadParty: the '%s' engine is not registered",
			models.PartySchemaName)
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.PartyFieldId:    partyId,
			models.PartyFieldOrgId: orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

// isArchived reads the archive flag, treating an absent one as "not archived".
//
// Absent means the column was never written, which is the state of every party created before
// archiving existed — reading that as archived would refuse references to most of the table.
func isArchived(record dmodel.DynamicFields) bool {
	archived := models.NewPartyFrom(record).IsArchived()
	return archived != nil && *archived
}
