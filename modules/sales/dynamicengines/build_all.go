package dynamicengines

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// BuildAllEngines resolves every registered onion, which is what runs its constructor and, through
// buildOnion, installs it into the resource hub.
//
// Registration alone builds nothing: deps records a constructor and dig constructs on first
// resolution. Without this pass a resource that no route and no sibling injects would never call
// InstallResource, and the first caller to reach it through the hub would fail at runtime --
// typically a cron sweep, hours after a boot that looked healthy.
//
// It runs as its own step at the end of the module's Init rather than inside each registration.
// That ordering is load-bearing: an onion whose closure injects one of Sales' own application
// services (the channel and the bill both take ChannelPaymentAppService) cannot be built until
// app.InitApplicationServices has registered it, which happens after InitDynamicEngines. Forcing
// at registration time fails the boot with "missing type: channel.ChannelPaymentAppService".
//
// A dig name tag is a compile-time struct tag, so this cannot loop over SchemaNames(); each
// resource is listed explicitly, and the count check in registry_test.go catches an omission.
func BuildAllEngines() error {
	return stdErr.Join(
		deps.Invoke(func(p salesFulfillmentMethodEngineParam) {}),
		deps.Invoke(func(p salesChannelEngineParam) {}),
		deps.Invoke(func(p salesPointEngineParam) {}),
		deps.Invoke(func(p salesPricelistEngineParam) {}),
		deps.Invoke(func(p salesPricelistItemEngineParam) {}),
		deps.Invoke(func(p salesComboEngineParam) {}),
		deps.Invoke(func(p salesComboComponentEngineParam) {}),
		deps.Invoke(func(p salesOrderEngineParam) {}),
		deps.Invoke(func(p salesOrderLineEngineParam) {}),
		deps.Invoke(func(p salesOrderLineComponentEngineParam) {}),
		deps.Invoke(func(p salesOrderAdjustmentEngineParam) {}),
		deps.Invoke(func(p salesOrderEventEngineParam) {}),
		deps.Invoke(func(p salesBillEngineParam) {}),
		deps.Invoke(func(p salesBillLineEngineParam) {}),
		deps.Invoke(func(p salesBillRelationEngineParam) {}),
		deps.Invoke(func(p salesPaymentEngineParam) {}),
		deps.Invoke(func(p salesFulfillmentRequestEngineParam) {}),
		deps.Invoke(func(p salesFulfillmentRequestLineEngineParam) {}),
		deps.Invoke(func(p salesOrderFulfillmentEngineParam) {}),
		deps.Invoke(func(p salesOrderFulfillmentItemEngineParam) {}),
		deps.Invoke(func(p salesFulfillmentAttemptEngineParam) {}),
		deps.Invoke(func(p salesFulfillmentAttemptItemEngineParam) {}),
		deps.Invoke(func(p salesFulfillmentTargetChangeEngineParam) {}),
		deps.Invoke(func(p salesReturnEngineParam) {}),
		deps.Invoke(func(p salesReturnLineEngineParam) {}),
		deps.Invoke(func(p salesRefundPaymentEngineParam) {}),
		deps.Invoke(func(p salesPromotionProgramEngineParam) {}),
		deps.Invoke(func(p salesPromotionConditionGroupEngineParam) {}),
		deps.Invoke(func(p salesPromotionConditionEngineParam) {}),
		deps.Invoke(func(p salesPromotionConditionTargetEngineParam) {}),
		deps.Invoke(func(p salesPromotionRewardEngineParam) {}),
		deps.Invoke(func(p salesPromotionCompatibilityEngineParam) {}),
		deps.Invoke(func(p salesVoucherCodeEngineParam) {}),
		deps.Invoke(func(p salesVoucherRedemptionEngineParam) {}),
		deps.Invoke(func(p salesQuotationEngineParam) {}),
		deps.Invoke(func(p salesQuotationLineEngineParam) {}),
		deps.Invoke(func(p salesFiscalRequestEngineParam) {}),
		deps.Invoke(func(p salesBillingInstructionEngineParam) {}),
		deps.Invoke(func(p salesBillingIssuanceAttemptEngineParam) {}),
		deps.Invoke(func(p salesManualDiscountEngineParam) {}),
		deps.Invoke(func(p salesIntegrationOutboxEngineParam) {}),
	)
}
