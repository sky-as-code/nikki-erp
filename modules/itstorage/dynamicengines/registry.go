// Package dynamicengines declares the resource onions IT Storage serves through the composable
// resource engine, and registers them and their typed layers into the dependency container.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func InitDynamicEngines() error {
	return stdErr.Join(
		registerSpaceEngine(),
		registerConsumptionEngine(),
		registerFileEngine(),
	)
}

// SpaceRestActions are the built-ins with a route: spaces are system-managed, so clients only read.
func SpaceRestActions() []dynamicresource.CrudAction {
	return []dynamicresource.CrudAction{
		dynamicresource.CrudActionGetById, dynamicresource.CrudActionSearch, dynamicresource.CrudActionExists,
		dynamicresource.CrudActionGetSchema,
	}
}

// ConsumptionRestActions are the read built-ins: the counters are kept by the upload and delete.
func ConsumptionRestActions() []dynamicresource.CrudAction {
	return SpaceRestActions()
}

// FileRestActions are the built-ins with a route. A file is created by the streamed upload only,
// and its metadata is immutable, so the generic create and update have no route.
func FileRestActions() []dynamicresource.CrudAction {
	return append(SpaceRestActions(), dynamicresource.CrudActionDelete)
}
