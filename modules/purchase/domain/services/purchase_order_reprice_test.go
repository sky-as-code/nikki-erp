package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Which orders may be repriced. The refusal rules are the pure decision in this operation;
// everything else it does is a write.

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

// A confirmed order is a document the vendor holds a copy of; moving its prices afterwards would
// make the two disagree with nothing on the record to say why.
func TestAConfirmedOrderIsNotRepriceable(t *testing.T) {
	status := string(models.PurchaseOrderStatusPurchaseOrder)

	refusal := repriceRefusal(orderWith(status, false), status)

	require.NotNil(t, refusal)
	require.Equal(t, 1, refusal.Count())
	assert.Equal(t, "purchase_order.not_repriceable", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, status,
		"the refusal must name the status, or the buyer cannot tell which rule they hit")
}

// A cancelled order gets its own message rather than the generic one, because the useful next step
// is to duplicate it.
func TestACancelledOrderIsRefusedWithItsOwnMessage(t *testing.T) {
	status := string(models.PurchaseOrderStatusCancelled)

	refusal := repriceRefusal(orderWith(status, false), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.cancelled_is_final", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, "duplicate")
}

// Locking is a separate axis from status, so a locked draft is refused too: repricing is an edit,
// and without this check locking would close an order to every edit except the one that changes
// every price on it.
func TestALockedDraftIsRefused(t *testing.T) {
	status := string(models.PurchaseOrderStatusRfq)

	refusal := repriceRefusal(orderWith(status, true), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.locked", (*refusal)[0].Key)
	assert.Contains(t, (*refusal)[0].Message, "unlock")
}

// Committed is checked before locked, so a confirmed and locked order reports the reason that will
// still be true after unlocking it.
func TestTheCommittedRefusalOutranksTheLockedOne(t *testing.T) {
	status := string(models.PurchaseOrderStatusPurchaseOrder)

	refusal := repriceRefusal(orderWith(status, true), status)

	require.NotNil(t, refusal)
	assert.Equal(t, "purchase_order.not_repriceable", (*refusal)[0].Key,
		"unlocking would not make a confirmed order repriceable, so that is not the advice to give")
}

// With no ports bound there is nothing to resolve against, so the operation reports that it
// examined nothing rather than failing. Only a hand-built service reaches this.
func TestRepricingWithoutPortsExaminesNothing(t *testing.T) {
	service := &PurchaseOrderDomainServiceImpl{}

	report, err := service.repriceLines(nil, orderWith(string(models.PurchaseOrderStatusRfq), false))

	require.NoError(t, err)
	assert.Zero(t, report.LinesExamined)
	assert.Empty(t, report.Changed)
}
