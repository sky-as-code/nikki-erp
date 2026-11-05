package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The draft order's promotion actions; each answers the PromotionQuote.

func (this *SalesOrderRest) RemoveVoucher(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "remove voucher", payload,
		this.orderSvc.RemoveVoucher, dynamicresource.Identity[any])
}

func (this *SalesOrderRest) SelectPromotionReward(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "select promotion reward", payload,
		this.orderSvc.SelectPromotionReward, dynamicresource.Identity[any])
}

func (this *SalesOrderRest) PreviewPromotions(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "preview promotions", payload,
		this.orderSvc.PreviewPromotions, dynamicresource.Identity[any])
}
