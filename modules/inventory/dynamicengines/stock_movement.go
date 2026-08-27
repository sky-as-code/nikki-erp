package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// The movement resources. Transfer carries the six operations that make stock move; move and move
// line are mostly read surfaces, because what writes them is the reservation and validation engine
// rather than a client.

// Permission codes for the movement operations.
//
// They are separate actions rather than folded into "update" because they are materially different
// powers: validating a transfer moves real goods and cannot be undone by an edit, while updating
// one changes a note. A role that may do the second should not thereby be able to do the first.
const (
	PermissionConfirm           = "confirm"
	PermissionCheckAvailability = "check_availability"
	PermissionReserve           = "reserve"
	PermissionUnreserve         = "unreserve"
	PermissionValidate          = "validate"
	PermissionCancel            = "cancel"
	// Raising a return commits the company to taking goods back, which is a commercial decision
	// rather than an edit to a shipping document, so it carries its own permission.
	PermissionCreateReturn = "create_return"
)

// Action names, namespaced by resource in the same style as the built-ins.
const (
	ActionConfirm           = "confirm"
	ActionCheckAvailability = "check_availability"
	ActionReserve           = "reserve"
	ActionUnreserve         = "unreserve"
	ActionValidate          = "validate"
	ActionCancel            = "cancel"
	ActionCreateReturn      = "create_return"
)

// Param names the movement actions read from the request.
const (
	paramTransferId      = "id"
	paramIdempotencyKey  = "idempotency_key"
	paramCreateBackorder = "create_backorder"
	paramReturnLines     = "lines"
	paramReturnMoveId    = "move_id"
	paramReturnQuantity  = "quantity"
)

func stockTransferEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockTransferSchemaName,
		DefineActions: defineStockTransferActions,
	}
}

func stockMoveEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockMoveSchemaName,
	}
}

func stockMoveLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockMoveLineSchemaName,
		DefineActions: defineStockMoveLineActions,
	}
}

func stockMoveDependencyEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockMoveDependencySchemaName,
	}
}

// defineStockMoveLineActions closes the move line's write surface.
//
// A move line is an allocation decision, not a document: the reservation engine creates it when it
// claims stock and validate stamps it when the stock moves. A client-written line would be a claim
// on a balance that the balance itself does not know about, so the two would disagree with nothing
// to reconcile them. Editing an allocation by hand needs the release-and-re-reserve flow of
// BR §4.2.5.4, which is not in this phase.
//
// As with the quant, the actions are refused rather than removed, so a caller gets a 400 naming the
// reason instead of a 404 that reads as a wrong URL.
func defineStockMoveLineActions(engine drif.DynamicResourceEngine) error {
	for _, action := range []string{drif.ActionCreate, drif.ActionUpdate, drif.ActionDelete} {
		err := engine.ModifyAction(drif.DynamicActionDelta{
			ActionName:    action,
			ValidateExtra: rejectMoveLineWrite,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to attach the stock move line '%s' guard", action)
		}
	}
	return nil
}

func rejectMoveLineWrite(
	_ corectx.Context, _ *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockMoveLineSchemaName,
		"stock_move_line.not_client_writable",
		"stock move lines are written by the reservation engine; reserve, unreserve or validate the "+
			"transfer instead",
	))
	return nil
}

// defineStockTransferActions exposes the six movement operations as engine actions.
//
// They are engine actions rather than hand-written REST handlers, per docs/wiki/07 §6.7: the engine
// already does the permission check, the param binding and the response shaping, and a handler
// would have to restate all three. Each is a POST, because none of them is a CRUD verb — a
// validate is not an update to a transfer, it is an event that happens to one.
func defineStockTransferActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfirm,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/confirm",
			Permission:  PermissionConfirm,
			MainProcess: processConfirm,
		}),
		// Read permission, because it takes nothing and changes nothing: it answers a question
		// about stock the caller can already see.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCheckAvailability,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/check_availability",
			Permission:  drif.PermissionRead,
			MainProcess: processCheckAvailability,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReserve,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/reserve",
			Permission:  PermissionReserve,
			MainProcess: processReserve,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionUnreserve,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/unreserve",
			Permission:  PermissionUnreserve,
			MainProcess: processUnreserve,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionValidate,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/validate",
			Permission:  PermissionValidate,
			MainProcess: processValidate,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancel,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancel,
			MainProcess: processCancel,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateReturn,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/create_return",
			Permission:  PermissionCreateReturn,
			MainProcess: processCreateReturn,
		}),
	)
}

func processCreateReturn(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}

	request, vErrs := readReturnRequest(input.Params)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := service.CreateReturn(ctx, readActionId(input), request)
	return toMutateActionResult(result, err)
}

// readReturnRequest reads the optional per-line quantities from the request body.
//
// Absent `lines` means "return everything still returnable" (F9), which is what the contextual
// action on the transfer page sends: the return lands as a draft and the user adjusts its move
// quantities through ordinary CRUD before confirming. An API caller that wants a partial return
// names the moves explicitly.
func readReturnRequest(params dmodel.DynamicFields) (services.ReturnRequest, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	request := services.ReturnRequest{}

	raw, present := params[paramReturnLines]
	if !present || raw == nil {
		return request, vErrs
	}

	items, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.lines_malformed",
			"'lines' must be a list of {move_id, quantity} entries"))
		return request, vErrs
	}

	for _, item := range items {
		line, vErr := readReturnLine(item)
		if vErr != nil {
			vErrs.Append(*vErr)
			continue
		}
		request.Lines = append(request.Lines, line)
	}
	return request, vErrs
}

func readReturnLine(item any) (services.ReturnLineRequest, *ft.ClientErrorItem) {
	fields, ok := item.(map[string]any)
	if !ok {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_malformed",
			"each return line must be an object with move_id and quantity")
	}

	moveId, _ := fields[paramReturnMoveId].(string)
	if moveId == "" {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_move_id_required",
			"each return line must name a move_id")
	}

	quantity, vErrs := readDecimalField(fields, paramReturnQuantity)
	if vErrs.Count() > 0 {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_quantity_malformed",
			"return line for move '"+moveId+"' must carry a decimal quantity")
	}
	return services.ReturnLineRequest{MoveId: moveId, Quantity: quantity}, nil
}

// transferServiceOf reaches the derived service the module installed during Init.
//
// The type assertion is what makes the extra operations reachable: the engine hands the action its
// service as the base interface, and only the derived type carries Confirm, Reserve and the rest.
// A failed assertion means Init did not install it, which is a wiring bug rather than a request
// problem.
func transferServiceOf(input drif.ProcessInput) (*services.StockTransferDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.StockTransferDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the stock transfer engine is not running the derived transfer service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func processConfirm(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Confirm(ctx, readActionId(input))
	return toMutateActionResult(result, err)
}

func processCheckAvailability(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	return service.CheckAvailability(ctx, readActionId(input))
}

func processReserve(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Reserve(ctx, readActionId(input))
	return toMutateActionResult(result, err)
}

func processUnreserve(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Unreserve(ctx, readActionId(input))
	return toMutateActionResult(result, err)
}

func processValidate(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Validate(
		ctx, readActionId(input), readStringField(input.Params, paramIdempotencyKey),
		readOptionalBool(input.Params, paramCreateBackorder))
	return toMutateActionResult(result, err)
}

func processCancel(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Cancel(ctx, readActionId(input))
	return toMutateActionResult(result, err)
}

func readActionId(input drif.ProcessInput) string {
	return readStringField(input.Params, paramTransferId)
}

// toMutateActionResult widens a mutation result into the engine's generic action result.
//
// The engine's own toActionResult is package-private to modules/dynamicresource/engine, so this is
// a local equivalent rather than an unnecessary duplicate: the ClientErrors must survive, because
// a refused operation reports its reason through them and not through err.
func toMutateActionResult(
	result *dyn.OpResult[dyn.MutateResultData], err error,
) (*drif.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	out := &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}

func readStringField(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

// readOptionalBool distinguishes "absent" from "false", which the `ask` backorder policy depends
// on: absent means the caller has not decided, and false means they decided not to.
func readOptionalBool(params dmodel.DynamicFields, field string) *bool {
	value, ok := params[field]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case bool:
		return &typed
	case *bool:
		return typed
	default:
		return nil
	}
}
