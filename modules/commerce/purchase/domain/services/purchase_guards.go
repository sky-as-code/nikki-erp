package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The write and delete rules of the Purchase resources, chained into the domain services' own
// Create/Update/Delete overrides. They run inside corecrud, so a direct service call is held to
// them exactly as a REST call is.

// An order and an agreement are deletable only from statuses where nothing has been committed to a
// counterparty. Everywhere else the correct operation is cancel, which keeps the record and its
// audit trail: a document a vendor has already seen cannot be made never to have existed.

// deletableOrderStatuses: an order may be removed only once cancelled. "rfq" is deliberately absent
// even though nothing has been sent — a draft still carries an allocated code, and cancelling it is
// cheap and leaves the trail.
var deletableOrderStatuses = map[string]bool{
	string(models.PurchaseOrderStatusCancelled): true,
}

// deletableAgreementStatuses: draft or cancelled. A draft agreement is deletable where a draft
// order is not, because an agreement's code is not quoted to a vendor until it is confirmed.
var deletableAgreementStatuses = map[string]bool{
	string(models.AgreementStatusDraft):     true,
	string(models.AgreementStatusCancelled): true,
}

func (this *PurchaseOrderDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.DeleteOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(GuardOrderDelete, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func (this *PurchaseAgreementDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.DeleteOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(GuardAgreementDelete, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

// GuardOrderDelete refuses to delete an order that is not cancelled. found is the stored record.
func GuardOrderDelete(
	_ corectx.Context, _ *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	assertDeletableStatus(
		found.GetFieldData(), models.PurchaseOrderFieldStatus, deletableOrderStatuses,
		"purchase_order.not_deletable",
		"only a cancelled purchase order can be deleted; cancel it instead",
		vErrs)
	return nil
}

// GuardAgreementDelete refuses to delete an agreement that is neither a draft nor cancelled.
func GuardAgreementDelete(
	_ corectx.Context, _ *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	assertDeletableStatus(
		found.GetFieldData(), models.AgreementFieldStatus, deletableAgreementStatuses,
		"purchase_agreement.not_deletable",
		"only a draft or cancelled purchase agreement can be deleted; cancel or close it instead",
		vErrs)
	return nil
}

// assertDeletableStatus refuses the delete unless the record's status is in the allowed set. An
// unreadable status is refused rather than allowed, so the guard fails closed.
func assertDeletableStatus(
	foundModel dmodel.DynamicFields, field string, allowed map[string]bool,
	errorKey string, message string, vErrs *ft.ClientErrors,
) {
	status := util.ValueOrZeroOf(foundModel.GetString(field))
	if allowed[status] {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey(errorKey, "purchase"), message))
}

// A vendor price is validated as it will be, not as it was sent: an update names only changed
// fields, so checking the payload alone would let a price change slip past the unit compatibility
// rule whose units are absent from the request.

func (this *VendorProductPriceDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateVendorPrice)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *VendorProductPriceDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateVendorPrice)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

func (this *VendorProductPriceDomainServiceImpl) validateVendorPrice(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if this.validator == nil || input == nil {
		return nil
	}
	record := input.GetFieldData()
	if found != nil {
		record = MergedRecord(found.GetFieldData(), input.GetFieldData())
	}
	return this.validator.Validate(ctx, record, vErrs)
}

// MergedRecord overlays the change on the stored record. It copies rather than mutating, because
// corecrud writes the input on its way out and validation must not be a side effect on it.
func MergedRecord(stored dmodel.DynamicFields, change dmodel.DynamicFields) dmodel.DynamicFields {
	merged := make(dmodel.DynamicFields, len(stored)+len(change))
	for field, value := range stored {
		merged[field] = value
	}
	for field, value := range change {
		merged[field] = value
	}
	return merged
}
