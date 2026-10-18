// Package dynamicengines declares the resource onions the IAM module serves through the
// composable resource engine, and registers them into the dependency container during the
// module's Init().
//
// Each resource file wires the module's own repository, domain service and application service
// onto the composable defaults and publishes those typed layers. Neither iam/app, iam/domain nor
// iam/infra may import this package, so both iam (which registers the onions) and
// iam/transport/restful (which registers their routes) can import it without a cycle.
package dynamicengines

import (
	stdErr "errors"
)

// InitDynamicEngines registers every resource onion this module owns. Nothing is built here: an
// onion is a container constructor, resolved the first time a consumer asks for it.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerUserEngine(),
		registerOrgEngine(),
		registerOrgUnitEngine(),
		registerGroupEngine(),
		registerEntitlementEngine(),
		registerRoleEngine(),
		registerResourceEngine(),
		registerActionEngine(),
		registerGrantRequestEngine(),
	)
}
