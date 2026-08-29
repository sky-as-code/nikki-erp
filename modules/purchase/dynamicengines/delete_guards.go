package dynamicengines

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

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

func defineOrderDeleteGuard(engine drif.DynamicResourceEngine) error {
	return engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   orderKeysToFetch,
		ValidateExtra: guardOrderDelete,
	})
}

func guardOrderDelete(
	_ corectx.Context, _ *drif.DynamicEntity, found *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	foundModel := found.GetFieldData()
	assertDeletableStatus(
		foundModel, models.PurchaseOrderFieldStatus, deletableOrderStatuses,
		"purchase_order.not_deletable",
		"only a cancelled purchase order can be deleted; cancel it instead",
		vErrs)
	return nil
}

func defineAgreementDeleteGuard(engine drif.DynamicResourceEngine) error {
	return engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   agreementKeysToFetch,
		ValidateExtra: guardAgreementDelete,
	})
}

func guardAgreementDelete(
	_ corectx.Context, _ *drif.DynamicEntity, found *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	foundModel := found.GetFieldData()
	assertDeletableStatus(
		foundModel, models.AgreementFieldStatus, deletableAgreementStatuses,
		"purchase_agreement.not_deletable",
		"only a draft or cancelled purchase agreement can be deleted; cancel or close it instead",
		vErrs)
	return nil
}

// The engine fetches the record named by these keys and hands it to the guard as foundModel, which
// is how the guard gets a status without a read of its own.
func orderKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.PurchaseOrderFieldId: params[models.PurchaseOrderFieldId]}
}

func agreementKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.AgreementFieldId: params[models.AgreementFieldId]}
}

// assertDeletableStatus refuses the delete unless the record's status is in the allowed set. An
// unreadable status is refused rather than allowed, so the guard fails closed.
func assertDeletableStatus(
	foundModel dmodel.DynamicFields,
	field string,
	allowed map[string]bool,
	errorKey string,
	message string,
	vErrs *ft.ClientErrors,
) {
	status := util.ValueOrZeroOf(foundModel.GetString(field))
	if allowed[status] {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(field, errorKey, message))
}
