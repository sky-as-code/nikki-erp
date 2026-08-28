package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Archiving and unarchiving a vendor price (section 25).
//
// The two directions are NOT symmetric, and that asymmetry is the whole of this file:
//
//   - ARCHIVING is always allowed. Withdrawing an offer is the vendor's or the buyer's prerogative
//     and nothing downstream breaks, because a purchase order line records the price it resolved
//     rather than a live reference to this row (PRICE-INV-024). An archived quote stays readable
//     forever; what it stops doing is pricing anything new.
//   - UNARCHIVING revalidates, because the world moved on while the row was retired. The vendor may
//     have been blocked, the product delisted, the unit archived. Bringing back a row that would be
//     refused if it were written today would put a price into resolution that no create or update
//     would have accepted — a validation rule that can be bypassed by archiving and unarchiving is
//     not a rule.
//
// It is a derived service rather than a ValidateExtra hook because the engine REFUSES that hook on
// set_archived: DefineAction reports that "the crud helper behind it accepts none, so the hook would
// never run". Attaching one would have failed at boot rather than silently — but it still means the
// check has to live here.

// NewVendorProductPriceDomainService derives the vendor price service from the engine's default.
//
// Only SetArchived is overridden. Create and update are already guarded through ValidateExtra,
// which the engine does support for them, and duplicating those checks here would give two places
// for the same rule to drift apart.
func NewVendorProductPriceDomainService(
	base drif.DynamicResourceService, validator *VendorPriceValidator,
) *VendorProductPriceDomainServiceImpl {
	return &VendorProductPriceDomainServiceImpl{
		DynamicResourceService: base,
		validator:              validator,
	}
}

type VendorProductPriceDomainServiceImpl struct {
	drif.DynamicResourceService

	// validator is the same one the write guards use, so archiving and creating enforce one set of
	// rules rather than two implementations of the same intent.
	validator *VendorPriceValidator
}

var _ drif.DynamicResourceService = (*VendorProductPriceDomainServiceImpl)(nil)

// SetArchived archives freely and revalidates on the way back (section 25).
func (this *VendorProductPriceDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if !isUnarchiving(params) {
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	stored, err := this.loadVendorPrice(ctx, stringOf(params, basemodel.FieldId))
	if err != nil {
		return nil, err
	}
	if stored == nil {
		// Not found is the base call's answer to give, in its own shape. Producing a second one
		// here would mean two different responses for the same condition depending on which
		// direction was asked for.
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	vErrs := ft.NewClientErrors()
	if this.validator != nil {
		if err := this.validator.Validate(ctx, stored, vErrs); err != nil {
			return nil, err
		}
	}
	if vErrs.Count() > 0 {
		// The refusals name the vendor, the product or the unit that has since become unusable,
		// which is what somebody restoring a price needs to know: the row is fine, the world
		// around it changed.
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	return this.DynamicResourceService.SetArchived(ctx, params)
}

// isUnarchiving reports whether this request is bringing a row back.
//
// An ABSENT is_archived is treated as archiving, matching the engine's own reading: the command
// carries a nil flag and the repository leaves the value alone, so nothing is being restored and
// there is nothing to revalidate. Only an explicit false asks for the check.
func isUnarchiving(params dmodel.DynamicFields) bool {
	value, present := params[basemodel.FieldIsArchived]
	if !present || value == nil {
		return false
	}
	archived, ok := value.(bool)
	if !ok {
		if pointer, isPointer := value.(*bool); isPointer && pointer != nil {
			archived = *pointer
		} else {
			return false
		}
	}
	return !archived
}

// loadVendorPrice reads the stored row so it can be validated as it stands.
//
// The stored record is validated, not the request: a set_archived request carries an id, an etag
// and a flag, and validating that would check nothing at all. What matters is whether the row's own
// vendor, product and unit are still usable.
func (this *VendorProductPriceDomainServiceImpl) loadVendorPrice(
	ctx corectx.Context, recordId string,
) (dmodel.DynamicFields, error) {
	if recordId == "" {
		return nil, nil
	}

	engine, err := engineFor(models.VendorProductPriceSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.VendorProductPriceFieldId: recordId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadVendorPrice")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}
