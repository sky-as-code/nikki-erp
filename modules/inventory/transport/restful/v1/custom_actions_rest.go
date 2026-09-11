package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The custom routes of the Inventory resources. Each handler hands the bound payload to the
// application service, which authorizes and delegates; the route itself is declared in
// transport/restful/index.go with RestEngine.AddRoute.

func (this *WarehouseRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "suspend warehouse", payload, this.warehouseSvc.Suspend, composable.MutateResponse)
}

func (this *WarehouseRest) Resume(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "resume warehouse", payload, this.warehouseSvc.Resume, composable.MutateResponse)
}

func (this *WarehouseRest) ConfigureIncomingFlow(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "configure incoming flow", payload,
		this.warehouseSvc.ConfigureIncomingFlowAction, composable.MutateResponse)
}

func (this *WarehouseRest) ConfigureOutgoingFlow(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "configure outgoing flow", payload,
		this.warehouseSvc.ConfigureOutgoingFlowAction, composable.MutateResponse)
}

func (this *InventoryLocationRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "suspend location", payload, this.inventoryLocationSvc.Suspend, composable.MutateResponse)
}

func (this *InventoryLocationRest) Resume(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "resume location", payload, this.inventoryLocationSvc.Resume, composable.MutateResponse)
}

func (this *InventoryLocationRest) Move(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "move location", payload, this.inventoryLocationSvc.Move, composable.MutateResponse)
}

func (this *PutawayRuleRest) SuggestLocation(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "suggest putaway location", payload, this.putawayRuleSvc.SuggestLocation, composable.Identity[any])
}

func (this *ProductTemplateRest) GenerateVariants(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "generate variants", payload, this.productTemplateSvc.GenerateVariants, composable.Identity[any])
}

func (this *ProductTemplateRest) ResolveSelection(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "resolve selection", payload, this.productTemplateSvc.ResolveSelection, composable.Identity[any])
}

func (this *ProductVariantRest) GetEffective(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "get effective product", payload, this.productVariantSvc.GetEffective, composable.Identity[any])
}

func (this *StockScrapRest) DoScrap(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "do scrap", payload, this.stockScrapSvc.DoScrap, composable.MutateResponse)
}

func (this *StockQuantRest) EnterCount(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "enter count", payload, this.stockQuantSvc.EnterCount, composable.MutateResponse)
}

func (this *StockQuantRest) ResetCount(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reset count", payload, this.stockQuantSvc.ResetCount, composable.MutateResponse)
}

func (this *StockQuantRest) ApplyAdjustment(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "apply adjustment", payload, this.stockQuantSvc.ApplyAdjustment, composable.MutateResponse)
}

func (this *StockQuantRest) ScheduleCount(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "schedule count", payload, this.stockQuantSvc.ScheduleCount, composable.MutateResponse)
}

func (this *StockQuantRest) AssignCounter(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "assign counter", payload, this.stockQuantSvc.AssignCounter, composable.MutateResponse)
}

func (this *StockQuantRest) VariantStockSummary(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "variant stock summary", payload, this.stockQuantSvc.VariantStockSummary, composable.Identity[any])
}

func (this *StockQuantRest) VariantStockSummaries(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "variant stock summaries", payload, this.stockQuantSvc.VariantStockSummaries, composable.Identity[any])
}

func (this *StockQuantRest) TemplateStockSummary(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "template stock summary", payload, this.stockQuantSvc.TemplateStockSummary, composable.Identity[any])
}

func (this *StockQuantRest) StockByWarehouse(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "stock by warehouse", payload, this.stockQuantSvc.StockByWarehouse, composable.Identity[any])
}

func (this *StockQuantRest) StockByLocation(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "stock by location", payload, this.stockQuantSvc.StockByLocation, composable.Identity[any])
}

func (this *StockQuantRest) ProductUsage(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "product usage", payload, this.stockQuantSvc.ProductUsage, composable.Identity[any])
}

func (this *StockTransferRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "confirm transfer", payload, this.stockTransferSvc.Confirm, composable.MutateResponse)
}

func (this *StockTransferRest) CheckAvailability(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "check availability", payload, this.stockTransferSvc.CheckAvailability, composable.Identity[any])
}

func (this *StockTransferRest) Reserve(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reserve transfer", payload, this.stockTransferSvc.Reserve, composable.MutateResponse)
}

func (this *StockTransferRest) Unreserve(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "unreserve transfer", payload, this.stockTransferSvc.Unreserve, composable.MutateResponse)
}

func (this *StockTransferRest) Validate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "validate transfer", payload, this.stockTransferSvc.Validate, composable.MutateResponse)
}

func (this *StockTransferRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel transfer", payload, this.stockTransferSvc.Cancel, composable.MutateResponse)
}

func (this *StockTransferRest) CreateReturn(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create return", payload, this.stockTransferSvc.CreateReturn, composable.MutateResponse)
}

func (this *StockTransferRest) ReserveForSource(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reserve for source", payload, this.stockTransferSvc.ReserveForSource, composable.Identity[any])
}

func (this *StockTransferRest) ReleaseForSource(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "release for source", payload, this.stockTransferSvc.ReleaseForSource, composable.MutateResponse)
}

func (this *StockTransferRest) ReallocateReservation(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reallocate reservation", payload, this.stockTransferSvc.ReallocateReservation, composable.Identity[any])
}

func (this *StockTransferRest) ApplyFulfillmentResult(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "apply fulfillment result", payload, this.stockTransferSvc.ApplyFulfillmentResult, composable.Identity[any])
}
