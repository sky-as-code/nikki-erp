package app

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

// InitApplicationServices registers what the application layer contributes to the container.
//
// Every resource application service is built by its resource onion in dynamicengines/, because
// each one needs the composable default the onion constructs. The settings update is the only
// service registered here: it belongs to no resource.
func InitApplicationServices() error {
	return deps.Register(NewSettingsApplicationServiceImpl)
}
