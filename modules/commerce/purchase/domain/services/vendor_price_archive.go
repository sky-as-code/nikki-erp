package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"

	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
)

// Archiving and unarchiving a vendor price. The two directions are not symmetric: archiving is
// always allowed, since an order line records the price it resolved rather than a live reference to
// this row, so an archived quote stays readable and only stops pricing anything new. Unarchiving
// revalidates, because the vendor, product or unit may have become unusable meanwhile, and a rule
// that archiving and unarchiving can bypass is not a rule. This is a derived service rather than a
// ValidateExtra hook because corecrud accepts no hook on set_archived.

// NewVendorProductPriceDomainService derives the vendor price service from the onion's default.
// Create and update chain the validator through ValidateExtra (purchase_guards.go).
func NewVendorProductPriceDomainService(
	base composable.CrudDomainService, validator *VendorPriceValidator,
) *VendorProductPriceDomainServiceImpl {
	return &VendorProductPriceDomainServiceImpl{
		CrudDomainService: base,
		validator:         validator,
	}
}

type VendorProductPriceDomainServiceImpl struct {
	composable.CrudDomainService

	// The same validator the write guards use, so archiving and creating enforce one set of rules.
	validator *VendorPriceValidator
}

// SetArchived archives freely and revalidates on the way back.
func (this *VendorProductPriceDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if !isUnarchiving(params) {
		return this.CrudDomainService.SetArchived(ctx, params)
	}

	stored, err := this.loadVendorPrice(ctx, stringOf(params, basemodel.FieldId))
	if err != nil {
		return nil, err
	}
	if stored == nil {
		// Let the base call report not-found in its own shape, so the answer does not depend on
		// which direction was asked for.
		return this.CrudDomainService.SetArchived(ctx, params)
	}

	vErrs := ft.NewClientErrors()
	if this.validator != nil {
		if err := this.validator.Validate(ctx, stored, vErrs); err != nil {
			return nil, err
		}
	}
	if vErrs.Count() > 0 {
		// The refusals name whichever of the vendor, product or unit has since become unusable.
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	return this.CrudDomainService.SetArchived(ctx, params)
}

// isUnarchiving reports whether this request is bringing a row back. An absent is_archived counts
// as archiving, matching the engine: the repository leaves the value alone, so nothing is restored
// and there is nothing to revalidate. Only an explicit false asks for the check.
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

// loadVendorPrice reads the stored row so it can be validated as it stands. The stored record is
// validated, not the request, which carries only an id, an etag and a flag.
func (this *VendorProductPriceDomainServiceImpl) loadVendorPrice(
	ctx corectx.Context, recordId string,
) (dmodel.DynamicFields, error) {
	if recordId == "" {
		return nil, nil
	}

	repo, err := repoFor(models.VendorProductPriceSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{
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
