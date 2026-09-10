// Package dynamicengines declares the resource onions the Essential module serves through the
// composable resource engine, and registers them into the dependency container during the
// module's Init().
//
// Each resource file wires the module's own repository, domain service and application service
// onto the composable defaults, and publishes those typed layers so that transport and other
// modules inject them by type. Nothing here is built eagerly: an onion is a container
// constructor, resolved the first time a consumer asks for it.
package dynamicengines

import (
	stdErr "errors"
)

// InitDynamicEngines registers every resource onion this module owns.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerCurrencyEngine(),
		registerUomEngine(),
		registerUomCatEngine(),
	)
}
