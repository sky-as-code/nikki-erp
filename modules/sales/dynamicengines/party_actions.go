package dynamicengines

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Naming the parties to a sale.
//
// Four actions over one service rather than four rule sets: `assign_parties` sets any combination in
// one call, and the three single-role actions exist because that is how a till and a back office
// actually work — one decision at a time, each with its own permission. A business that may record
// who bought is not thereby allowed to redirect the invoice.
//
// Every one is all-or-nothing. A combined call refused on bill-to writes neither sold-to nor payer,
// so an order never ends up half-assigned by a request that reported a failure.

const (
	PermissionAssignParties = "assign_parties"
	PermissionAssignSoldTo  = "assign_sold_to_party"
	PermissionAssignBillTo  = "assign_bill_to_party"
	PermissionAssignPayer   = "assign_payer_party"

	ActionAssignParties = "assign_parties"
	ActionAssignSoldTo  = "assign_sold_to_party"
	ActionAssignBillTo  = "assign_bill_to_party"
	ActionAssignPayer   = "assign_payer_party"
)

const (
	paramSoldToPartyId = "sold_to_party_id"
	paramPayerPartyId  = "payer_party_id"
)

// partyPort is the bound contacts party port. A package variable because an action callback is
// handed only its own engine and cannot reach the container.
var partyPort itExt.PartyExtService

// SetPartyPort must be called by Init before any request is served.
func SetPartyPort(port itExt.PartyExtService) {
	partyPort = port
}

func defineSalesOrderPartyActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAssignParties,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/assign_parties",
			Permission:  PermissionAssignParties,
			MainProcess: processAssignParties,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAssignSoldTo,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/assign_sold_to_party",
			Permission:  PermissionAssignSoldTo,
			MainProcess: processAssignSoldTo,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAssignBillTo,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/assign_bill_to_party",
			Permission:  PermissionAssignBillTo,
			MainProcess: processAssignBillTo,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAssignPayer,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/assign_payer_party",
			Permission:  PermissionAssignPayer,
			MainProcess: processAssignPayer,
		}),
	)
}

// processAssignParties sets any combination of the three roles in one call.
//
// A role the caller did not mention is left alone; one sent as an explicit null is cleared. That
// distinction is why the params are read by presence rather than by value.
func processAssignParties(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	return runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(input.Params, paramId),
		SoldTo:       readPartyAssignment(input.Params, paramSoldToPartyId),
		BillTo:       readPartyAssignment(input.Params, paramBillToPartyId),
		Payer:        readPartyAssignment(input.Params, paramPayerPartyId),
	})
}

func processAssignSoldTo(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	return runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(input.Params, paramId),
		SoldTo:       readPartyAssignment(input.Params, paramSoldToPartyId),
	})
}

func processAssignBillTo(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	return runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(input.Params, paramId),
		BillTo:       readPartyAssignment(input.Params, paramBillToPartyId),
	})
}

func processAssignPayer(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	return runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(input.Params, paramId),
		Payer:        readPartyAssignment(input.Params, paramPayerPartyId),
	})
}

func runPartyAssignment(
	ctx corectx.Context, params services.AssignPartiesParams,
) (*drif.ActionResult, error) {
	vErrs, err := services.AssignParties(ctx, params, partyPort)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{HasData: false}, nil
}

// readPartyAssignment reads one role, distinguishing "not mentioned" from "explicitly cleared".
//
// A plain string reader cannot tell those apart — both arrive as empty — and collapsing them would
// mean either that a caller can never clear a role, or that omitting one silently wipes it. Presence
// in the map is the signal, exactly as the API contract describes it (CR §24.1).
func readPartyAssignment(params dmodel.DynamicFields, field string) services.PartyAssignment {
	value, present := params[field]
	if !present {
		return services.PartyAssignment{}
	}
	if value == nil {
		return services.PartyAssignment{Requested: true}
	}
	return services.PartyAssignment{
		Requested: true,
		PartyId:   readStringParam(params, field),
	}
}
