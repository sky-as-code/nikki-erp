package v1

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The custom routes of the Inventory resources. Each handler hands the bound payload to the
// application service, which authorizes and delegates; the route itself is declared in
// transport/restful/index.go with RestEngine.AddRoute.

func (this *WarehouseRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "suspend warehouse", payload, this.warehouseSvc.Suspend, dynamicresource.MutateResponse)
}

func (this *WarehouseRest) Resume(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(
		echoCtx, "resume warehouse", payload, this.warehouseSvc.Resume, dynamicresource.MutateResponse,
	)
}

func (this *WarehouseRest) ConfigureIncomingFlow(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "configure incoming flow", payload,
		this.warehouseSvc.ConfigureIncomingFlowAction, dynamicresource.MutateResponse)
}

func (this *WarehouseRest) ConfigureOutgoingFlow(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "configure outgoing flow", payload,
		this.warehouseSvc.ConfigureOutgoingFlowAction, dynamicresource.MutateResponse)
}

func (this *InventoryLocationRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "suspend location", payload, this.inventoryLocationSvc.Suspend, dynamicresource.MutateResponse)
}

func (this *InventoryLocationRest) Resume(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "resume location", payload, this.inventoryLocationSvc.Resume, dynamicresource.MutateResponse)
}

func (this *InventoryLocationRest) Move(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "move location", payload, this.inventoryLocationSvc.Move, dynamicresource.MutateResponse)
}

func (this *PutawayRuleRest) SuggestLocation(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "suggest putaway location", payload, this.putawayRuleSvc.SuggestLocation, dynamicresource.Identity[any])
}

func (this *ProductTemplateRest) GenerateVariants(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "generate variants", payload, this.productTemplateSvc.GenerateVariants, dynamicresource.Identity[any])
}

func (this *ProductTemplateRest) ResolveSelection(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "resolve selection", payload, this.productTemplateSvc.ResolveSelection, dynamicresource.Identity[any])
}

// Duplicate answers like the built-in create: the new template's id and etag.
func (this *ProductTemplateRest) Duplicate(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "duplicate product template", payload, this.productTemplateSvc.Duplicate,
		func(created dmodel.DynamicFields) any { return httpserver.NewRestCreateResponseDyn(created) })
}

func (this *ProductTemplateRest) PreviewAttributeConfiguration(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "preview attribute configuration", payload,
		this.productTemplateSvc.PreviewAttributeConfiguration, dynamicresource.Identity[any])
}

func (this *ProductVariantRest) GetEffective(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "get effective product", payload, this.productVariantSvc.GetEffective, dynamicresource.Identity[any])
}

func (this *StockScrapRest) DoScrap(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(
		echoCtx, "do scrap", payload, this.stockScrapSvc.DoScrap, dynamicresource.MutateResponse,
	)
}

func (this *StockQuantRest) EnterCount(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(
		echoCtx, "enter count", payload, this.stockQuantSvc.EnterCount, dynamicresource.MutateResponse,
	)
}

func (this *StockQuantRest) ResetCount(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(
		echoCtx, "reset count", payload, this.stockQuantSvc.ResetCount, dynamicresource.MutateResponse,
	)
}

func (this *StockQuantRest) ApplyAdjustment(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "apply adjustment", payload, this.stockQuantSvc.ApplyAdjustment, dynamicresource.MutateResponse)
}

func (this *StockQuantRest) ScheduleCount(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "schedule count", payload, this.stockQuantSvc.ScheduleCount, dynamicresource.MutateResponse)
}

func (this *StockQuantRest) AssignCounter(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "assign counter", payload, this.stockQuantSvc.AssignCounter, dynamicresource.MutateResponse)
}

func (this *StockQuantRest) VariantStockSummary(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "variant stock summary", payload, this.stockQuantSvc.VariantStockSummary, dynamicresource.Identity[any])
}

func (this *StockQuantRest) VariantStockSummaries(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "variant stock summaries", payload, this.stockQuantSvc.VariantStockSummaries, dynamicresource.Identity[any])
}

func (this *StockQuantRest) TemplateStockSummary(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "template stock summary", payload, this.stockQuantSvc.TemplateStockSummary, dynamicresource.Identity[any])
}

func (this *StockQuantRest) StockByWarehouse(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "stock by warehouse", payload, this.stockQuantSvc.StockByWarehouse, dynamicresource.Identity[any])
}

func (this *StockQuantRest) StockByLocation(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "stock by location", payload, this.stockQuantSvc.StockByLocation, dynamicresource.Identity[any])
}

func (this *StockQuantRest) ProductUsage(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "product usage", payload, this.stockQuantSvc.ProductUsage, dynamicresource.Identity[any])
}

func (this *StockTransferRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "confirm transfer", payload, this.stockTransferSvc.Confirm, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) CheckAvailability(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "check availability", payload, this.stockTransferSvc.CheckAvailability, dynamicresource.Identity[any])
}

func (this *StockTransferRest) Reserve(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "reserve transfer", payload, this.stockTransferSvc.Reserve, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) Unreserve(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "unreserve transfer", payload, this.stockTransferSvc.Unreserve, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) Validate(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "validate transfer", payload, this.stockTransferSvc.Validate, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "cancel transfer", payload, this.stockTransferSvc.Cancel, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) CreateReturn(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "create return", payload, this.stockTransferSvc.CreateReturn, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) ReserveForSource(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "reserve for source", payload, this.stockTransferSvc.ReserveForSource, dynamicresource.Identity[any])
}

func (this *StockTransferRest) ReleaseForSource(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "release for source", payload, this.stockTransferSvc.ReleaseForSource, dynamicresource.MutateResponse)
}

func (this *StockTransferRest) ReallocateReservation(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "reallocate reservation", payload, this.stockTransferSvc.ReallocateReservation, dynamicresource.Identity[any])
}

func (this *StockTransferRest) ApplyFulfillmentResult(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "apply fulfillment result", payload, this.stockTransferSvc.ApplyFulfillmentResult, dynamicresource.Identity[any])
}

func (this *StockReservationRest) ReserveWarehouseStock(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "reserve warehouse stock", payload, this.stockReservationSvc.ReserveWarehouseStock, dynamicresource.Identity[any])
}

func (this *StockReservationRest) CheckWarehouseAvailability(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "check warehouse availability", payload, this.stockReservationSvc.CheckWarehouseAvailability, dynamicresource.Identity[any])
}

func (this *StockReservationRest) ConsumeReservation(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "consume reservation", payload, this.stockReservationSvc.ConsumeReservation, dynamicresource.Identity[any])
}

func (this *StockReservationRest) ReleaseReservation(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "release reservation", payload, this.stockReservationSvc.ReleaseReservation, dynamicresource.Identity[any])
}
