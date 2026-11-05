package restful

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The voucher routes (plan §3.2): CRUD with the built-in archive, the collection-level
// generate_batch and the record-level revoke.

func initSalesVoucherV1(route *echo.Group) error {
	engines, err := dynamicresource.ResolveEngines(models.SalesVoucherSchemaName)
	if err != nil {
		return err
	}
	rest := &voucherRest{voucherSvc: engines[0].ApplicationService().(itPromotion.SalesVoucherApplicationService)}
	rest.SetApplicationService(engines[0].ApplicationService())
	return dynamicresource.NewRestEngine(models.SalesVoucherSchemaName, rest).
		AddRoute(dynamicresource.RouteDefinition{
			Path: "generate_batch", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.GenerateBatch,
		}).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/revoke", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Revoke,
		}).
		RegisterRoutes(route)
}

type voucherRest struct {
	dynamicresource.CrudRestBase
	voucherSvc itPromotion.SalesVoucherApplicationService
}

func (this *voucherRest) GenerateBatch(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "generate sales vouchers", payload,
		this.voucherSvc.GenerateBatch, dynamicresource.Identity[map[string]any])
}

func (this *voucherRest) Revoke(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "revoke sales voucher", payload,
		this.voucherSvc.Revoke, dynamicresource.Identity[dmodel.DynamicFields])
}
