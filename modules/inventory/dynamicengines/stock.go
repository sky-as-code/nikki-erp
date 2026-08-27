package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// Stock Operation Type is plain CRUD master data: everything it needs is already expressed by its
// schema. Stock Quant is not — it is current state rather than a document, so its engine takes its
// write actions away.
//
// Inventory Location used to live here too. It moved to inventory_location.go when it became the
// module's shared location master rather than a stock-owned resource.

func stockOperationTypeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockOperationTypeSchemaName,
	}
}

func stockQuantEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockQuantSchemaName,
		DefineActions: defineStockQuantActions,
	}
}

// defineStockQuantActions closes the quant's write surface.
//
// A balance is not something a client sets; it is the running total of the movements that have
// completed against it. Leaving create, update and delete open would allow an on-hand quantity
// with no movement behind it, which no report could explain and no audit could trace. Corrections
// go through an inventory adjustment, a transfer or a scrap. See BR §3.3, §4.2.2.6, AC-STOCK-002.
//
// The actions are refused rather than removed so that a caller gets a 400 naming the reason,
// instead of a 404 that reads as "wrong URL".
//
// TODO: move to higher validation layer.
func defineStockQuantActions(engine drif.DynamicResourceEngine) error {
	// for _, action := range []string{drif.ActionCreate, drif.ActionUpdate, drif.ActionDelete} {
	// 	err := engine.ModifyAction(drif.DynamicActionDelta{
	// 		ActionName:    action,
	// 		ValidateExtra: rejectQuantWrite,
	// 	})
	// 	if err != nil {
	// 		return errors.Wrapf(err, "failed to attach the stock quant '%s' guard", action)
	// 	}
	// }
	if err := defineStockCountActions(engine); err != nil {
		return err
	}
	// The product-facing reads live here too: what they read is quants, and putting them on the
	// product engines would have Product owning a stock query. See product_stock_actions.go.
	return defineProductStockActions(engine)
}

// Permission codes for the counting operations.
//
// They are separate actions rather than folded into one, because they are different powers held by
// different people: a warehouse hand enters what they counted, and a supervisor decides that the
// count becomes the balance. Applying an adjustment writes stock that no trade movement explains,
// which is the most sensitive thing this module lets anyone do.
const (
	PermissionEnterCount      = "enter_count"
	PermissionResetCount      = "reset_count"
	PermissionApplyAdjustment = "apply_adjustment"
	PermissionScheduleCount   = "schedule_count"
	PermissionAssignCounter   = "assign_counter"
)

const (
	ActionEnterCount      = "enter_count"
	ActionResetCount      = "reset_count"
	ActionApplyAdjustment = "apply_adjustment"
	ActionScheduleCount   = "schedule_count"
	ActionAssignCounter   = "assign_counter"
)

// Param names the counting actions read from the request body.
const (
	paramQuantId         = "id"
	paramCountedQuantity = "counted_quantity"
	paramCountReasonCode = "count_reason_code"
	paramCountReasonText = "count_reason_text"
	paramNextCountDate   = "next_count_date"
	paramAssignedUserId  = "count_assigned_user_id"
)

// defineStockCountActions adds physical inventory and cycle counting to the quant.
//
// These write to a resource whose create, update and delete are all refused just above, which
// looks contradictory and is not: the guard exists to stop a client setting a balance with no
// movement behind it, and none of these actions touches on_hand_quantity. Enter and Reset write
// only count metadata; Apply changes the balance solely by generating a movement. See BR §4.2.7.
func defineStockCountActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionEnterCount,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/enter_count",
			Permission:  PermissionEnterCount,
			MainProcess: processEnterCount,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionResetCount,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/reset_count",
			Permission:  PermissionResetCount,
			MainProcess: processResetCount,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionApplyAdjustment,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/apply_adjustment",
			Permission:  PermissionApplyAdjustment,
			MainProcess: processApplyAdjustment,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionScheduleCount,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/schedule_count",
			Permission:  PermissionScheduleCount,
			MainProcess: processScheduleCount,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAssignCounter,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/assign_counter",
			Permission:  PermissionAssignCounter,
			MainProcess: processAssignCounter,
		}),
	)
}

// quantServiceOf reaches the derived service the module installed during Init.
//
// A failed assertion means Init did not install it, which is a wiring bug rather than a request
// problem — the same reasoning as transferServiceOf.
func quantServiceOf(input drif.ProcessInput) (*services.StockQuantDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.StockQuantDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the stock quant engine is not running the derived quant service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func processEnterCount(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := quantServiceOf(input)
	if err != nil {
		return nil, err
	}

	counted, vErrs := readDecimalField(input.Params, paramCountedQuantity)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := service.EnterCount(
		ctx, readStringField(input.Params, paramQuantId), counted,
		readStringField(input.Params, paramCountReasonCode),
		readStringField(input.Params, paramCountReasonText))
	return toMutateActionResult(result, err)
}

func processResetCount(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := quantServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.ResetCount(ctx, readStringField(input.Params, paramQuantId))
	return toMutateActionResult(result, err)
}

func processApplyAdjustment(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := quantServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.ApplyAdjustment(ctx, readStringField(input.Params, paramQuantId))
	return toMutateActionResult(result, err)
}

func processScheduleCount(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := quantServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.ScheduleCount(
		ctx, readStringField(input.Params, paramQuantId),
		readStringField(input.Params, paramNextCountDate))
	return toMutateActionResult(result, err)
}

func processAssignCounter(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := quantServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.AssignCounter(
		ctx, readStringField(input.Params, paramQuantId),
		readStringField(input.Params, paramAssignedUserId))
	return toMutateActionResult(result, err)
}

// readDecimalField reads a quantity from the request body.
//
// Quantities travel as strings (BR §7.3) so that a float's rounding never reaches a balance, but a
// JSON client may still send a bare number. Both are accepted and parsed through decimal, which is
// the only representation the rest of the module works in.
func readDecimalField(params dmodel.DynamicFields, field string) (decimal.Decimal, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	value, ok := params[field]
	if !ok || value == nil {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant."+field+"_required",
			"'"+field+"' is required"))
		return decimal.Zero, vErrs
	}

	switch typed := value.(type) {
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			vErrs.Append(*ft.NewBusinessViolation(
				models.StockQuantSchemaName, "stock_quant."+field+"_malformed",
				"'"+field+"' must be a decimal number"))
			return decimal.Zero, vErrs
		}
		return parsed, vErrs
	case float64:
		return decimal.NewFromFloat(typed), vErrs
	case int:
		return decimal.NewFromInt(int64(typed)), vErrs
	case int64:
		return decimal.NewFromInt(typed), vErrs
	default:
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant."+field+"_malformed",
			"'"+field+"' must be a decimal number"))
		return decimal.Zero, vErrs
	}
}

func rejectQuantWrite(
	_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	services.AssertQuantNotClientWritable(vErrs)
	return nil
}
