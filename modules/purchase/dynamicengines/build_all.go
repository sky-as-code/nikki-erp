package dynamicengines

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// BuildAllEngines resolves every registered onion, which runs its constructor and installs it into
// the resource hub. Registration alone builds nothing, and a resource no route injects would never
// be installed: the first lifecycle action or usage check reaching it through the hub would fail.
//
// A dig name tag is a compile-time struct tag, so each resource is listed explicitly; the boot
// check in the module's Init catches an omission.
func BuildAllEngines() error {
	return stdErr.Join(
		deps.Invoke(func(p configurationEngineParam) {}),
		deps.Invoke(func(p vendorProductPriceEngineParam) {}),
		deps.Invoke(func(p sourcingGroupEngineParam) {}),
		deps.Invoke(func(p auditEventEngineParam) {}),
		deps.Invoke(func(p agreementEngineParam) {}),
		deps.Invoke(func(p agreementLineEngineParam) {}),
		deps.Invoke(func(p purchaseOrderEngineParam) {}),
		deps.Invoke(func(p purchaseOrderLineEngineParam) {}),
	)
}
