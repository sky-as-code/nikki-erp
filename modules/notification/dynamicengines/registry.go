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

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/services"
	infraExt "github.com/sky-as-code/nikki-erp/modules/notification/infra/external"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// InitDynamicEngines registers every resource onion this module owns.
//
// Order is irrelevant to registration, but not to resolution: the notification onion injects the
// recipient and delivery repositories, so those two must be registered for its constructor to
// resolve. Registering all three here means they always are.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerSendNormalizer(),
		registerRecipientEngine(),
		registerDeliveryEngine(),
		registerNotificationEngine(),
	)
}

// registerSendNormalizer binds the send rules onto the attached channels.
//
// It is here rather than beside the channels themselves because this is the one package that
// already knows both layers: the rules are domain, the channels are infrastructure, and having the
// infrastructure import the domain to wire them would point the dependency the wrong way.
func registerSendNormalizer() error {
	return deps.Register(
		func(dispatcher *infraExt.ChannelDispatcher) *services.SendNormalizer {
			return services.NewSendNormalizer(dispatcher)
		},

		// The fan-out reaches the channels through the same dispatcher the rules do, so that a
		// channel accepted when a notification is sent is the one that delivers it.
		func(
			dispatcher *infraExt.ChannelDispatcher,
			deliveries it.DeliveryRepository,
			logger logging.LoggerService,
		) *app.ChannelFanOut {
			return app.NewChannelFanOut(dispatcher, deliveries, logger)
		},
	)
}
