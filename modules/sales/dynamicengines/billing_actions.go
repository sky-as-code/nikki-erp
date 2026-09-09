package dynamicengines

import (
	stdErr "errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// The billing instruction's lifecycle, as actions rather than field updates.
//
// Marking ready takes its own permission because it is a materially different power from correcting
// a typo: it releases a legal document to be issued in a company's name. A role that may fix a
// mistyped tax code must not thereby be able to bill someone.

const (
	PermissionCreateBillingInstruction = "create"
	PermissionUpdateBillingInstruction = "update"
	PermissionMarkBillingReady         = "mark_ready"
	PermissionRevertBillingToDraft     = "revert_to_draft"
	PermissionCancelBillingInstruction = "cancel"

	ActionCreateBillingInstruction = "create_billing_instruction"
	ActionUpdateBillingInstruction = "update_billing_instruction"
	ActionMarkBillingReady         = "mark_ready"
	ActionRevertBillingToDraft     = "revert_to_draft"
	ActionCancelBillingInstruction = "cancel"
)

const (
	paramSalesOrderId   = "sales_order_id"
	paramBillToPartyId  = "bill_to_party_id"
	paramTaxId          = "tax_id"
	paramLegalName      = "legal_name"
	paramBillingAddress = "billing_address"
	paramBillingEmail   = "billing_email"
	paramBillingSource  = "source"

	// The checkbox on the release screen: re-read the buyer from their party record rather than
	// billing the snapshot already captured. Only meaningful on a retry.
	paramFetchLatestPartyDetails = "fetch_latest_party_details"
)

func defineSalesBillingInstructionActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		// Collection-level: creating names the order in the body, since there is no instruction to
		// address yet.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateBillingInstruction,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "create_billing_instruction",
			Permission:  PermissionCreateBillingInstruction,
			MainProcess: processCreateBillingInstruction,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionUpdateBillingInstruction,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/update_billing_instruction",
			Permission:  PermissionUpdateBillingInstruction,
			MainProcess: processUpdateBillingInstruction,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionMarkBillingReady,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/mark_ready",
			Permission:  PermissionMarkBillingReady,
			MainProcess: processMarkBillingReady,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionRevertBillingToDraft,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/revert_to_draft",
			Permission:  PermissionRevertBillingToDraft,
			MainProcess: processRevertBillingToDraft,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancelBillingInstruction,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancelBillingInstruction,
			MainProcess: processCancelBillingInstruction,
		}),
	)
}

func processCreateBillingInstruction(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	instructionId, vErrs, err := services.CreateBillingInstruction(ctx,
		services.CreateBillingInstructionParams{
			SalesOrderId:  readStringParam(input.Params, paramSalesOrderId),
			BillToPartyId: readStringParam(input.Params, paramBillToPartyId),
			Source:        readStringParam(input.Params, paramBillingSource),
			Snapshot:      readBillingSnapshot(input),
		})
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{HasData: true, Data: map[string]any{
		"sales_billing_instruction_id": instructionId,
	}}, nil
}

func processUpdateBillingInstruction(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	vErrs, err := services.UpdateBillingInstructionSnapshot(ctx,
		readStringParam(input.Params, paramId), readBillingSnapshot(input))
	return billingActionResult(vErrs, err)
}

func processMarkBillingReady(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	vErrs, err := services.MarkBillingInstructionReady(ctx,
		readStringParam(input.Params, paramId),
		readBoolParam(input.Params, paramFetchLatestPartyDetails),
		partyPort)
	return billingActionResult(vErrs, err)
}

func processRevertBillingToDraft(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	vErrs, err := services.RevertBillingInstructionToDraft(ctx,
		readStringParam(input.Params, paramId))
	return billingActionResult(vErrs, err)
}

func processCancelBillingInstruction(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	vErrs, err := services.CancelBillingInstruction(ctx,
		readStringParam(input.Params, paramId))
	return billingActionResult(vErrs, err)
}

// readBillingSnapshot pulls the four confirmed fiscal fields off the request.
func readBillingSnapshot(input drif.ProcessInput) services.BillingSnapshot {
	return services.BillingSnapshot{
		TaxId:          readStringParam(input.Params, paramTaxId),
		LegalName:      readStringParam(input.Params, paramLegalName),
		BillingAddress: readStringParam(input.Params, paramBillingAddress),
		BillingEmail:   readStringParam(input.Params, paramBillingEmail),
	}
}

// billingActionResult is the shape every lifecycle action returns: the transition either happened or
// was refused for a reason the caller can act on.
func billingActionResult(vErrs *ft.ClientErrors, err error) (*drif.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{HasData: true, Data: map[string]any{"updated": true}}, nil
}
