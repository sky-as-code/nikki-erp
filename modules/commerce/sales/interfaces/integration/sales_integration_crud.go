// Package integration declares the layers of the two records Sales keeps about itself: the manual
// discounts granted on an order, and the outbox of integration events other systems consume.
//
// Both are read-only. A manual discount is written by the order action that granted it, so a
// writable row could forge a discount nobody authorized; an outbox row is a published event, and
// rewriting one would change history a consumer has already replayed.
package integration

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SalesManualDiscountRepository interface {
	composable.CrudRepository
}

type SalesManualDiscountApplicationService interface {
	composable.CrudApplicationService
}

type SalesIntegrationOutboxRepository interface {
	composable.CrudRepository
}

type SalesIntegrationOutboxApplicationService interface {
	composable.CrudApplicationService
}
