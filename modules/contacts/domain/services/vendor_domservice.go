package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewVendorDomainService is handed the composable default by the vendor profile onion. The CRUD
// is the default's; what this type adds is the capability a purchasing module asks of a vendor.
func NewVendorDomainService(base composable.CrudDomainService) itVendor.VendorDomainService {
	return &VendorDomainServiceImpl{CrudDomainService: base}
}

type VendorDomainServiceImpl struct {
	composable.CrudDomainService
}

// GetVendor fetches a party's vendor profile, so that a purchasing module can validate a vendor
// reference and read its defaults without reaching into Contacts' repositories.
func (this *VendorDomainServiceImpl) GetVendor(
	ctx corectx.Context, query itVendor.GetVendorQuery,
) (*itVendor.GetVendorResult, error) {
	profile, err := this.loadProfile(ctx, query.PartyId, query.OrgId)
	if err != nil {
		return nil, errors.Wrap(err, "get vendor")
	}
	if profile == nil {
		// Absence is a legitimate answer: the party exists but is not a vendor of this
		// organization. "Is a vendor" is defined as "has a profile row", so this IS the check.
		return &itVendor.GetVendorResult{HasData: false}, nil
	}

	return &itVendor.GetVendorResult{
		Data:    toVendorResultData(profile),
		HasData: true,
	}, nil
}

// AssertOrderable reports whether a new order may name this party as its vendor.
//
// The two failure cases are separate violations because they are different situations for whoever
// reads the error: a party nobody has qualified as a supplier, versus one that was qualified and is
// now suspended, blacklisted or archived.
func (this *VendorDomainServiceImpl) AssertOrderable(
	ctx corectx.Context, query itVendor.AssertOrderableQuery,
) (*itVendor.AssertOrderableResult, error) {
	vErrs := &ft.ClientErrors{}
	field := query.Field
	if field == "" {
		field = models.VendorProfileFieldPartyId
	}

	profile, err := this.loadProfile(ctx, query.PartyId, query.OrgId)
	if err != nil {
		return nil, errors.Wrap(err, "assert vendor orderable")
	}

	switch {
	case profile == nil:
		vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("vendor.not_a_vendor", "contacts"),
			"the party has no vendor profile in this organization"))
	case !profile.IsOrderable():
		vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("vendor.not_orderable", "contacts"),
			"the vendor is not active"))
	}

	if vErrs.Count() > 0 {
		return &itVendor.AssertOrderableResult{ClientErrors: *vErrs}, nil
	}
	return &itVendor.AssertOrderableResult{HasData: true}, nil
}

// loadProfile fetches the vendor profile of a party within an organization, returning nil when the
// party is not a vendor there.
//
// Both ids are part of the filter because the profile is keyed by the pair: the same party may be a
// vendor of one organization and an ordinary contact of another.
func (this *VendorDomainServiceImpl) loadProfile(
	ctx corectx.Context, partyId model.Id, orgId model.Id,
) (*models.VendorProfile, error) {
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.VendorProfileFieldPartyId: partyId,
			models.VendorProfileFieldOrgId:   orgId,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadProfile")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewVendorProfileFrom(found.Data), nil
}

func toVendorResultData(src *models.VendorProfile) itVendor.GetVendorResultData {
	return itVendor.GetVendorResultData{
		PartyId:           util.ValueOrZeroOf(src.GetPartyId()),
		OrgId:             util.ValueOrZeroOf(src.GetOrgId()),
		Status:            util.ValueOrZeroOf(src.GetStatus()),
		IsOrderable:       src.IsOrderable(),
		DefaultCurrencyId: src.GetDefaultCurrencyId(),
		PaymentTerms:      util.ValueOrZeroOf(src.GetPaymentTerms()),
		LeadTimeDays:      src.GetLeadTimeDays(),
	}
}
