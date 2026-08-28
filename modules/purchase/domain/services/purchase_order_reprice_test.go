package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Which orders may be repriced (section 30). The refusal rules are the pure decision in this
// operation, and the one worth pinning: everything else it does is a write, but WHO may not do it
// is a business rule that would otherwise only be visible by trying.

func orderWith(status string, locked bool) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderFieldStatus:   status,
		models.PurchaseOrderFieldIsLocked: locked,
	}
}

// A draft is the whole point: prices move while the buyer is still preparing the order.
func TestADraftOrderMayBeRepriced(t *testing.T) {
	for _, status := range []string{
		string(models.PurchaseOrderStatusRfq),
		string(models.PurchaseOrderStatusRfqSent),
		string(models.PurchaseOrderStatusToApprove),
	} {
		t.Run(status, func(t *testing.T) {
			assert.Nil(t, repriceRefusal(orderWith(status, false), status))
		})
	}
}

// The rule the operation exists to protect. A confirmed order is a document the vendor holds a copy
// of; moving its prices afterwards would make the two disagree with nothing on the record to say
// why, which is exactly the silent rewrite section 30 forbids.
func TestAConfirmedOrderIsNotRepriceable(t *testing.T) {
	status := string(models.PurchaseOrderStatusPurchaseOrder)

	refusal := repriceRefusal(orderWith(status, false), status)

	require.NotNil(t, refusal)
	require.Equal(t, 1, refusal.Count())
	assert.Equal(t, "purchase_order.not_repriceable", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, status,
		"the refusal must name the status, or the buyer cannot tell which rule they hit")
}

// A cancelled order gets its own message rather than the generic one, for the same reason the
// transition rules give it one: "you cannot" is not an answer, and the useful next step is to
// duplicate it.
func TestACancelledOrderIsRefusedWithItsOwnMessage(t *testing.T) {
	status := string(models.PurchaseOrderStatusCancelled)

	refusal := repriceRefusal(orderWith(status, false), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.cancelled_is_final", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, "duplicate")
}

// Locking is a separate axis from status (PUR-R2), so a locked DRAFT is refused too. Repricing is
// an edit, and the fact that it edits many lines at once rather than one is not an exemption —
// without this check, locking an order would close it to every edit except the one that changes
// every price on it.
func TestALockedDraftIsRefused(t *testing.T) {
	status := string(models.PurchaseOrderStatusRfq)

	refusal := repriceRefusal(orderWith(status, true), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.locked", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, "unlock")
}

// Committed is checked before locked, so a confirmed AND locked order reports the reason that will
// still be true after unlocking it. Reporting "locked" first would send somebody to unlock an order
// that would then refuse them anyway.
func TestTheCommittedRefusalOutranksTheLockedOne(t *testing.T) {
	status := string(models.PurchaseOrderStatusPurchaseOrder)

	refusal := repriceRefusal(orderWith(status, true), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.not_repriceable", (*refusal)[0].Key,
		"unlocking would not make a confirmed order repriceable, so that is not the advice to give")
}

// With no ports bound there is nothing to resolve against, and the operation reports that it
// examined nothing rather than failing. Only a hand-built service reaches this.
func TestRepricingWithoutPortsExaminesNothing(t *testing.T) {
	service := &PurchaseOrderDomainServiceImpl{}

	report, err := service.repriceLines(nil, orderWith(string(models.PurchaseOrderStatusRfq), false))

	require.NoError(t, err)
	assert.Zero(t, report.LinesExamined)
	assert.Empty(t, report.Changed)
}
