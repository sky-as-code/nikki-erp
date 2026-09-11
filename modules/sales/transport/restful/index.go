package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/sales/transport/restful/v1"
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
	return stdErr.Join(err, initSalesV1())
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
		return composable.NewRestEngine(models.SalesManualDiscountSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesIntegrationOutboxV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesIntegrationOutboxRest) error {
		return composable.NewRestEngine(models.SalesIntegrationOutboxSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentMethodV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentMethodRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentMethodSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: ":id/archive", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/unarchive", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Unarchive,
			}).
			RegisterRoutes(route)
	})
}

// The collection-level resolve is added before the :id routes so the router cannot read "resolve"
// as an id.
func initSalesChannelV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesChannelRest) error {
		return composable.NewRestEngine(models.SalesChannelSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "resolve", ActionType: composable.ActionTypeRead, HandlerFn: rest.Resolve,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/suspend", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Suspend,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/activate", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Activate,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/archive", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/payment_methods", ActionType: composable.ActionTypeRead, HandlerFn: rest.PaymentMethods,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/enable_payment_method",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.EnablePaymentMethod,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/disable_payment_method",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.DisablePaymentMethod,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPointV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPointRest) error {
		return composable.NewRestEngine(models.SalesPointSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: ":id/suspend", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Suspend,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/activate", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Activate,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/archive", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Archive,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/unarchive", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Unarchive,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPricelistV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPricelistRest) error {
		return composable.NewRestEngine(models.SalesPricelistSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: ":id/set_default", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.SetDefault,
			}).
			RegisterRoutes(route)
	})
}

func initSalesPricelistItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPricelistItemRest) error {
		return composable.NewRestEngine(models.SalesPricelistItemSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesComboV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesComboRest) error {
		return composable.NewRestEngine(models.SalesComboSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesComboComponentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesComboComponentRest) error {
		return composable.NewRestEngine(models.SalesComboComponentSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// The order's custom routes. create_order and fulfillment_targets are collection-level and are
// added before the :id routes, so the router cannot read either path segment as an id.
func initSalesOrderV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderRest) error {
		return composable.NewRestEngine(models.SalesOrderSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "create_order", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateOrder,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       "fulfillment_targets/search",
				ActionType: composable.ActionTypeRead, HandlerFn: rest.SearchFulfillmentTargets,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/reprice", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Reprice,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/confirm", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Confirm,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/cancel", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/apply_voucher", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ApplyVoucher,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/explain_price", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ExplainPrice,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/manual_discount", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.GrantManualDiscount,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/revoke_manual_discount",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.RevokeManualDiscount,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/assign_parties", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.AssignParties,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/assign_sold_to_party",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.AssignSoldToParty,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/assign_bill_to_party",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.AssignBillToParty,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/assign_payer_party",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.AssignPayerParty,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/fulfillments", ActionType: composable.ActionTypeRead, HandlerFn: rest.ListFulfillments,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/refunds", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateRefunds,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/refunds_view", ActionType: composable.ActionTypeRead, HandlerFn: rest.ViewRefunds,
			}).
			RegisterRoutes(route)
	})
}

func initSalesOrderLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderLineRest) error {
		return composable.NewRestEngine(models.SalesOrderLineSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// The three read-only order resources: only the read half of the built-in table is registered, so
// a write is not merely denied by IAM but has no route at all.
func initSalesOrderLineComponentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderLineComponentRest) error {
		return composable.NewRestEngine(models.SalesOrderLineComponentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderAdjustmentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderAdjustmentRest) error {
		return composable.NewRestEngine(models.SalesOrderAdjustmentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderEventV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderEventRest) error {
		return composable.NewRestEngine(models.SalesOrderEventSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func readOnlyRoutes() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionGetById,
		composable.CrudActionSearch,
		composable.CrudActionExists,
		composable.CrudActionGetSchema,
	}
}

// The bill's actions. merge is collection-level and is added before the :id routes so the router
// cannot read "merge" as an id.
func initSalesBillV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillRest) error {
		return composable.NewRestEngine(models.SalesBillSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "merge", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Merge,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/split", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Split,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/pay", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Pay,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/settle", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Settle,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/start_gateway_payment",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.StartGatewayPayment,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillLineRest) error {
		return composable.NewRestEngine(models.SalesBillLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesBillRelationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillRelationRest) error {
		return composable.NewRestEngine(models.SalesBillRelationSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesPaymentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPaymentRest) error {
		return composable.NewRestEngine(models.SalesPaymentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentRequestV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentRequestRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentRequestSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentRequestLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentRequestLineRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentRequestLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesOrderFulfillmentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderFulfillmentRest) error {
		return composable.NewRestEngine(models.SalesOrderFulfillmentSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: ":id/attempts", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateAttempt,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/attempt_result",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ApplyAttemptResult,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/reassign_target",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReassignTarget,
			}).
			RegisterRoutes(route)
	})
}

func initSalesOrderFulfillmentItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesOrderFulfillmentItemRest) error {
		return composable.NewRestEngine(models.SalesOrderFulfillmentItemSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentAttemptV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentAttemptRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentAttemptSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentAttemptItemV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentAttemptItemRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentAttemptItemSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesFulfillmentTargetChangeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFulfillmentTargetChangeRest) error {
		return composable.NewRestEngine(models.SalesFulfillmentTargetChangeSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

// create_return is collection-level and is added before the :id routes.
func initSalesReturnV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesReturnRest) error {
		return composable.NewRestEngine(models.SalesReturnSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "create_return", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateReturn,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/process", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Process,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/cancel", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesReturnLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesReturnLineRest) error {
		return composable.NewRestEngine(models.SalesReturnLineSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesRefundPaymentV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesRefundPaymentRest) error {
		return composable.NewRestEngine(models.SalesRefundPaymentSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesPromotionProgramV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionProgramRest) error {
		return composable.NewRestEngine(models.SalesPromotionProgramSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionGroupV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionGroupRest) error {
		return composable.NewRestEngine(models.SalesPromotionConditionGroupSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionRest) error {
		return composable.NewRestEngine(models.SalesPromotionConditionSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionConditionTargetV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionConditionTargetRest) error {
		return composable.NewRestEngine(models.SalesPromotionConditionTargetSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionRewardV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionRewardRest) error {
		return composable.NewRestEngine(models.SalesPromotionRewardSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesPromotionCompatibilityV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionCompatibilityRest) error {
		return composable.NewRestEngine(models.SalesPromotionCompatibilitySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesVoucherCodeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesVoucherCodeRest) error {
		return composable.NewRestEngine(models.SalesVoucherCodeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSalesVoucherRedemptionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesVoucherRedemptionRest) error {
		return composable.NewRestEngine(models.SalesVoucherRedemptionSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}

func initSalesQuotationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesQuotationRest) error {
		return composable.NewRestEngine(models.SalesQuotationSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: ":id/convert", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Convert,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/send", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Send,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/cancel", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesQuotationLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesQuotationLineRest) error {
		return composable.NewRestEngine(models.SalesQuotationLineSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// request_invoice and create_billing_instruction are collection-level, so both are added before
// their resource's :id routes.
func initSalesFiscalRequestV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesFiscalRequestRest) error {
		return composable.NewRestEngine(models.SalesFiscalRequestSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "request_invoice", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.RequestInvoice,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillingInstructionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillingInstructionRest) error {
		return composable.NewRestEngine(models.SalesBillingInstructionSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path:       "create_billing_instruction",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateInstruction,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/update_billing_instruction",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.UpdateInstruction,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/mark_ready", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.MarkReady,
			}).
			AddRoute(composable.RouteDefinition{
				Path:       ":id/revert_to_draft",
				ActionType: composable.ActionTypeGeneric, HandlerFn: rest.RevertToDraft,
			}).
			AddRoute(composable.RouteDefinition{
				Path: ":id/cancel", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Cancel,
			}).
			RegisterRoutes(route)
	})
}

func initSalesBillingIssuanceAttemptV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesBillingIssuanceAttemptRest) error {
		return composable.NewRestEngine(models.SalesBillingIssuanceAttemptSchemaName, rest).
			AddCrudRoutes(readOnlyRoutes()...).
			RegisterRoutes(route)
	})
}
