package dynamicengines

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// A purchase order and an agreement are deletable only from the statuses where nothing has been
// committed to a counterparty yet. Everywhere else the correct operation is cancel, which leaves
// the record and its audit trail in place.
//
// The distinction is the point: deleting a confirmed order would remove the evidence that the
// business committed to a purchase, while cancelling records that it did and then stopped. A
// document a vendor has already seen cannot be made never to have existed.

// deletableOrderStatuses is the whole of it: an order may be removed only once cancelled (BR 24).
//
// Note "rfq" is deliberately NOT here even though nothing has been sent yet. A draft still carries
// a code that was allocated to it, and the requirement asks for one deletion rule rather than a
// status-by-status one; cancelling a draft is cheap and leaves the trail.
var deletableOrderStatuses = map[string]bool{
	string(models.PurchaseOrderStatusCancelled): true,
}

// deletableAgreementStatuses: draft or cancelled (BR 46). A draft agreement is deletable where a
// draft order is not, because an agreement's code is not quoted to a vendor until it is confirmed.
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
// is how the guard gets a status to test without issuing a read of its own.
func orderKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.PurchaseOrderFieldId: params[models.PurchaseOrderFieldId]}
}

func agreementKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.AgreementFieldId: params[models.AgreementFieldId]}
}

// assertDeletableStatus refuses the delete unless the record's status is in the allowed set.
//
// A record whose status cannot be read is refused rather than allowed. The engine fetches the
// record before this runs, so an absent status means something is wrong with the record or the
// fetch — and defaulting to "delete it" on an unreadable status would make the guard fail open,
// which is the one failure mode a guard must not have.
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
