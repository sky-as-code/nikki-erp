package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/commerce/sales/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// Every Sales resource declares its routes explicitly through the composable RestEngine: the
// built-in CRUD table, narrowed to the read half for the resources a client must not write, plus
// one AddRoute per custom action.

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewSalesFulfillmentMethodRest,
		v1.NewSalesChannelRest,
		v1.NewSalesPointRest,
		v1.NewSalesPricelistRest,
		v1.NewSalesPricelistItemRest,
		v1.NewSalesComboRest,
		v1.NewSalesComboComponentRest,
		v1.NewSalesOrderRest,
		v1.NewSalesOrderLineRest,
		v1.NewSalesOrderLineComponentRest,
		v1.NewSalesOrderAdjustmentRest,
		v1.NewSalesOrderEventRest,
		v1.NewSalesBillRest,
		v1.NewSalesBillLineRest,
		v1.NewSalesBillRelationRest,
		v1.NewSalesPaymentRest,
		v1.NewSalesFulfillmentRequestRest,
		v1.NewSalesFulfillmentRequestLineRest,
		v1.NewSalesOrderFulfillmentRest,
		v1.NewSalesOrderFulfillmentItemRest,
		v1.NewSalesFulfillmentAttemptRest,
		v1.NewSalesFulfillmentAttemptItemRest,
		v1.NewSalesFulfillmentTargetChangeRest,
		v1.NewSalesReturnRest,
		v1.NewSalesReturnLineRest,
		v1.NewSalesRefundPaymentRest,
		v1.NewSalesPromotionProgramRest,
		v1.NewSalesPromotionConditionGroupRest,
		v1.NewSalesPromotionConditionRest,
		v1.NewSalesPromotionConditionTargetRest,
		v1.NewSalesPromotionRewardRest,
		v1.NewSalesPromotionCompatibilityRest,
		v1.NewSalesVoucherCodeRest,
		v1.NewSalesVoucherRedemptionRest,
		v1.NewSalesQuotationRest,
		v1.NewSalesQuotationLineRest,
		v1.NewSalesFiscalRequestRest,
		v1.NewSalesBillingInstructionRest,
		v1.NewSalesBillingIssuanceAttemptRest,
		v1.NewSalesManualDiscountRest,
		v1.NewSalesIntegrationOutboxRest,
	)
	return stdErr.Join(err, initSalesV1(), initSettingsV1())
}

func initSalesV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group(modconstants.SalesRouteV1)
		return stdErr.Join(
			initSalesFulfillmentMethodV1(routeV1),
			initSalesChannelV1(routeV1),
			initSalesPointV1(routeV1),
			initSalesPricelistV1(routeV1),
			initSalesPricelistItemV1(routeV1),
			initSalesComboV1(routeV1),
			initSalesComboComponentV1(routeV1),
			initSalesOrderV1(routeV1),
			initSalesOrderLineV1(routeV1),
			initSalesOrderLineComponentV1(routeV1),
			initSalesOrderAdjustmentV1(routeV1),
			initSalesOrderEventV1(routeV1),
			initSalesBillV1(routeV1),
			initSalesBillLineV1(routeV1),
			initSalesBillRelationV1(routeV1),
			initSalesPaymentV1(routeV1),
			initSalesFulfillmentRequestV1(routeV1),
			initSalesFulfillmentRequestLineV1(routeV1),
			initSalesOrderFulfillmentV1(routeV1),
			initSalesOrderFulfillmentItemV1(routeV1),
			initSalesFulfillmentAttemptV1(routeV1),
			initSalesFulfillmentAttemptItemV1(routeV1),
			initSalesFulfillmentTargetChangeV1(routeV1),
			initSalesReturnV1(routeV1),
			initSalesReturnLineV1(routeV1),
			initSalesRefundPaymentV1(routeV1),
			initSalesPromotionProgramV1(routeV1),
			initSalesPromotionConditionGroupV1(routeV1),
			initSalesPromotionConditionV1(routeV1),
			initSalesPromotionConditionTargetV1(routeV1),
			initSalesPromotionRewardV1(routeV1),
			initSalesPromotionCompatibilityV1(routeV1),
			initSalesVoucherCodeV1(routeV1),
			initSalesVoucherRedemptionV1(routeV1),
			initSalesQuotationV1(routeV1),
			initSalesQuotationLineV1(routeV1),
			initSalesFiscalRequestV1(routeV1),
			initSalesBillingInstructionV1(routeV1),
			initSalesBillingIssuanceAttemptV1(routeV1),
			initSalesManualDiscountV1(routeV1),
			initSalesIntegrationOutboxV1(routeV1),
		)
	})
}

// Both are read-only records Sales keeps about itself, so only the read half of the built-in
// table is registered.
func initSalesManualDiscountV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesManualDiscountRest) error {
		return dynamicresource.NewRestEngine(models.SalesManualDiscountSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesIntegrationOutboxV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesIntegrationOutboxRest) error {
		return dynamicresource.NewRestEngine(models.SalesIntegrationOutboxSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentMethodV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentMethodRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentMethodSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/archive", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/unarchive", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Unarchive,
			}).
			RegisterRoutes(route)
	})
}

// The collection-level resolve is added before the :id routes so the router cannot read "resolve"
// as an id.
func initSalesChannelV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesChannelRest) error {
		return dynamicresource.NewRestEngine(models.SalesChannelSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: "resolve", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.Resolve,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/suspend", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Suspend,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/activate", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Activate,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/archive", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/payment_methods", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.PaymentMethods,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/enable_payment_method",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.EnablePaymentMethod,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/disable_payment_method",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.DisablePaymentMethod,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPointV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPointRest) error {
		return dynamicresource.NewRestEngine(models.SalesPointSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/suspend", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Suspend,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/activate", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Activate,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/archive", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/unarchive", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Unarchive,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/payment_methods", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.PaymentMethods,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/enable_payment_method",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.EnablePaymentMethod,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/disable_payment_method",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.DisablePaymentMethod,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPricelistV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPricelistRest) error {
		return dynamicresource.NewRestEngine(models.SalesPricelistSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/set_default", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.SetDefault,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPricelistItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPricelistItemRest) error {
		return dynamicresource.NewRestEngine(models.SalesPricelistItemSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesComboV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesComboRest) error {
		return dynamicresource.NewRestEngine(models.SalesComboSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesComboComponentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesComboComponentRest) error {
		return dynamicresource.NewRestEngine(models.SalesComboComponentSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// The order's custom routes. create_order and fulfillment_targets are collection-level and are
// added before the :id routes, so the router cannot read either path segment as an id.
func initSalesOrderV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: "create_order", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateOrder,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       "fulfillment_targets/search",
				ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.SearchFulfillmentTargets,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/reprice", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Reprice,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/confirm", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Confirm,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/cancel", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/apply_voucher", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ApplyVoucher,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/explain_price", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ExplainPrice,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/manual_discount", ActionType: dynamicresource.ActionTypeGeneric,
				HandlerFn: rest.GrantManualDiscount,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/revoke_manual_discount",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.RevokeManualDiscount,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/assign_parties", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.AssignParties,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/assign_sold_to_party",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.AssignSoldToParty,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/assign_bill_to_party",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.AssignBillToParty,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/assign_payer_party",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.AssignPayerParty,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/fulfillments", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.ListFulfillments,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/refunds", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateRefunds,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/refunds_view", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.ViewRefunds,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/bills", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.ListBills,
			}).
			RegisterRoutes(route)
	})
}

func initSalesOrderLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderLineRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderLineSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// The three read-only order resources: only the read half of the built-in table is registered, so
// a write is not merely denied by IAM but has no route at all.
func initSalesOrderLineComponentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderLineComponentRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderLineComponentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderAdjustmentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderAdjustmentRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderAdjustmentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderEventV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderEventRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderEventSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func readOnlyRoutes() []dynamicresource.CrudAction {
	return []dynamicresource.CrudAction{
		dynamicresource.CrudActionGetById,
		dynamicresource.CrudActionSearch,
		dynamicresource.CrudActionExists,
		dynamicresource.CrudActionGetSchema,
	}
}

// The bill's actions. merge is collection-level and is added before the :id routes so the router
// cannot read "merge" as an id.
func initSalesBillV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillRest) error {
		return dynamicresource.NewRestEngine(models.SalesBillSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: "merge", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Merge,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/split", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Split,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/pay", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Pay,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/settle", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Settle,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/start_gateway_payment",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.StartGatewayPayment,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillLineRest) error {
		return dynamicresource.NewRestEngine(models.SalesBillLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesBillRelationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillRelationRest) error {
		return dynamicresource.NewRestEngine(models.SalesBillRelationSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesPaymentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPaymentRest) error {
		return dynamicresource.NewRestEngine(models.SalesPaymentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentRequestV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentRequestRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentRequestSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentRequestLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentRequestLineRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentRequestLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderFulfillmentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderFulfillmentRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderFulfillmentSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/attempts", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateAttempt,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/attempt_result",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ApplyAttemptResult,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/reassign_target",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReassignTarget,
			}).
			RegisterRoutes(route)
	})
}

func initSalesOrderFulfillmentItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderFulfillmentItemRest) error {
		return dynamicresource.NewRestEngine(models.SalesOrderFulfillmentItemSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentAttemptV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentAttemptRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentAttemptSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentAttemptItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentAttemptItemRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentAttemptItemSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentTargetChangeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentTargetChangeRest) error {
		return dynamicresource.NewRestEngine(models.SalesFulfillmentTargetChangeSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

// create_return is collection-level and is added before the :id routes.
func initSalesReturnV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesReturnRest) error {
		return dynamicresource.NewRestEngine(models.SalesReturnSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: "create_return", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateReturn,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/confirm", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Confirm,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/process", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Process,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/cancel", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesReturnLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesReturnLineRest) error {
		return dynamicresource.NewRestEngine(models.SalesReturnLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesRefundPaymentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesRefundPaymentRest) error {
		return dynamicresource.NewRestEngine(models.SalesRefundPaymentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesPromotionProgramV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionProgramRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionProgramSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionGroupV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionGroupRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionConditionGroupSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionConditionSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionTargetV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionTargetRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionConditionTargetSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionRewardV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionRewardRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionRewardSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionCompatibilityV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionCompatibilityRest) error {
		return dynamicresource.NewRestEngine(models.SalesPromotionCompatibilitySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesVoucherCodeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesVoucherCodeRest) error {
		return dynamicresource.NewRestEngine(models.SalesVoucherCodeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesVoucherRedemptionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesVoucherRedemptionRest) error {
		return dynamicresource.NewRestEngine(models.SalesVoucherRedemptionSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesQuotationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesQuotationRest) error {
		return dynamicresource.NewRestEngine(models.SalesQuotationSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/convert", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Convert,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/send", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Send,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/cancel", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesQuotationLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesQuotationLineRest) error {
		return dynamicresource.NewRestEngine(models.SalesQuotationLineSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// request_invoice and create_billing_instruction are collection-level, so both are added before
// their resource's :id routes.
func initSalesFiscalRequestV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFiscalRequestRest) error {
		return dynamicresource.NewRestEngine(models.SalesFiscalRequestSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path: "request_invoice", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.RequestInvoice,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillingInstructionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillingInstructionRest) error {
		return dynamicresource.NewRestEngine(models.SalesBillingInstructionSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{
				Path:       "create_billing_instruction",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateInstruction,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/update_billing_instruction",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.UpdateInstruction,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/mark_ready", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.MarkReady,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path:       ":id/revert_to_draft",
				ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.RevertToDraft,
			}).
			AddRoute(dynamicresource.RouteDefinition{
				Path: ":id/cancel", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillingIssuanceAttemptV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillingIssuanceAttemptRest) error {
		return dynamicresource.NewRestEngine(models.SalesBillingIssuanceAttemptSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}
