package dynamicengines

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// BuildAllEngines resolves every registered onion, which runs its constructor and installs it into
// the resource hub. Registration alone builds nothing, and a resource no route injects would never
// be installed: the first sweep or webhook reaching it through the hub would fail at runtime.
//
// A dig name tag is a compile-time struct tag, so each resource is listed explicitly; the boot
// check in OnAppStarted catches an omission.
func BuildAllEngines() error {
	return stdErr.Join(
		deps.Invoke(func(p paymentMethodEngineParam) {}),
		deps.Invoke(func(p paymentProfileEngineParam) {}),
		deps.Invoke(func(p transactionEngineParam) {}),
		deps.Invoke(func(p orderEngineParam) {}),
		deps.Invoke(func(p invoiceEngineParam) {}),
		deps.Invoke(func(p invoiceLineEngineParam) {}),
	)
}
