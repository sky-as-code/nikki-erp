// Package app holds the Payment & Invoice application services. Authorization happens here: the
// onion application services assert the permission of every built-in and custom action before
// delegating to the domain. The onions themselves are registered by dynamicengines, which also
// publishes PaymentMethodAppService from the payment method's onion.
package app
