// Package infra binds the ports every part of Commerce consumes. Each part keeps its own
// infra/<part>/external, so splitting a part into its own process still changes one package.
package infra

import (
	stdErr "errors"

	invext "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/external"
	purext "github.com/sky-as-code/nikki-erp/modules/commerce/infra/purchase/external"
	salesext "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/external"
)

// InitSubModule registers every part's port bindings. It runs before any part builds its engines:
// a binding is a container provider, resolved only when a derived service is constructed, so
// registering all three first is safe and lets a part bind a port another part serves.
func InitSubModule() error {
	return stdErr.Join(
		invext.InitExternal(),
		purext.InitExternal(),
		salesext.InitExternal(),
	)
}
