// Package dynamicengines declares the resource onions the Notification module serves through the
// composable resource engine, and registers them into the dependency container during the module's
// Init().
//
// Each resource file wires the module's own repository, domain service and application service onto
// the composable defaults, and publishes those typed layers so that transport injects them by type.
// Nothing here is built eagerly: an onion is a container constructor, resolved the first time a
// consumer asks for it.
package dynamicengines

import (
	stdErr "errors"
)

// InitDynamicEngines registers every resource onion this module owns.
//
// Order is irrelevant to registration, but not to resolution: the notification onion injects the
// recipient and delivery repositories, so those two must be registered for its constructor to
// resolve. Registering all three here means they always are.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerRecipientEngine(),
		registerDeliveryEngine(),
		registerNotificationEngine(),
	)
}
