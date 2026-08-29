package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The order transition table.
func TestCanTransitionOrder(t *testing.T) {
	testCases := []struct {
		from, to models.PurchaseOrderStatus
		allowed  bool
	}{
		{models.PurchaseOrderStatusRfq, models.PurchaseOrderStatusRfqSent, true},
		{models.PurchaseOrderStatusRfq, models.PurchaseOrderStatusToApprove, true},
		{models.PurchaseOrderStatusRfq, models.PurchaseOrderStatusPurchaseOrder, true},
		{models.PurchaseOrderStatusRfq, models.PurchaseOrderStatusCancelled, true},

		{models.PurchaseOrderStatusRfqSent, models.PurchaseOrderStatusPurchaseOrder, true},
		{models.PurchaseOrderStatusRfqSent, models.PurchaseOrderStatusCancelled, true},
		// A sent quotation cannot go back to being unsent: the vendor has it.
		{models.PurchaseOrderStatusRfqSent, models.PurchaseOrderStatusRfq, false},

		{models.PurchaseOrderStatusToApprove, models.PurchaseOrderStatusPurchaseOrder, true},
		{models.PurchaseOrderStatusToApprove, models.PurchaseOrderStatusCancelled, true},
		// An order waiting on an approver cannot be pulled back into draft by its buyer.
		{models.PurchaseOrderStatusToApprove, models.PurchaseOrderStatusRfq, false},

		// A confirmed order can still be cancelled — a deal can fall through.
		{models.PurchaseOrderStatusPurchaseOrder, models.PurchaseOrderStatusCancelled, true},
		{models.PurchaseOrderStatusPurchaseOrder, models.PurchaseOrderStatusToApprove, false},
		{models.PurchaseOrderStatusPurchaseOrder, models.PurchaseOrderStatusRfq, false},

		// Cancelled is terminal in every direction.
		{models.PurchaseOrderStatusCancelled, models.PurchaseOrderStatusRfq, false},
		{models.PurchaseOrderStatusCancelled, models.PurchaseOrderStatusPurchaseOrder, false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.from)+" to "+string(testCase.to), func(t *testing.T) {
			assert.Equal(t, testCase.allowed,
				CanTransitionOrder(string(testCase.from), string(testCase.to)))
		})
	}
}

// A transition to the current status is allowed, so a retry after a lost response is not an error.
func TestATransitionToTheSameStatusIsAllowed(t *testing.T) {
	for status := range orderTransitions {
		assert.True(t, CanTransitionOrder(status, status), "%s to itself", status)
	}
	for status := range agreementTransitions {
		assert.True(t, CanTransitionAgreement(status, status), "%s to itself", status)
	}
}

func TestCanTransitionAgreement(t *testing.T) {
	testCases := []struct {
		from, to models.AgreementStatus
		allowed  bool
	}{
		{models.AgreementStatusDraft, models.AgreementStatusConfirmed, true},
		{models.AgreementStatusDraft, models.AgreementStatusCancelled, true},
		{models.AgreementStatusDraft, models.AgreementStatusClosed, false},
		{models.AgreementStatusConfirmed, models.AgreementStatusClosed, true},
		{models.AgreementStatusConfirmed, models.AgreementStatusCancelled, true},
		{models.AgreementStatusConfirmed, models.AgreementStatusDraft, false},
		{models.AgreementStatusClosed, models.AgreementStatusConfirmed, false},
		{models.AgreementStatusCancelled, models.AgreementStatusDraft, false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.from)+" to "+string(testCase.to), func(t *testing.T) {
			assert.Equal(t, testCase.allowed,
				CanTransitionAgreement(string(testCase.from), string(testCase.to)))
		})
	}
}

// Every status the table names must be one the schema declares; a typo is a row nothing can reach,
// which no compiler catches.
func TestTransitionTablesUseDeclaredStatuses(t *testing.T) {
	orderStatuses := map[string]bool{
		string(models.PurchaseOrderStatusRfq):           true,
		string(models.PurchaseOrderStatusRfqSent):       true,
		string(models.PurchaseOrderStatusToApprove):     true,
		string(models.PurchaseOrderStatusPurchaseOrder): true,
		string(models.PurchaseOrderStatusCancelled):     true,
	}
	for from, targets := range orderTransitions {
		assert.True(t, orderStatuses[from], "unknown order status %q", from)
		for _, to := range targets {
			assert.True(t, orderStatuses[to], "unknown order status %q", to)
		}
	}
	// Every declared status must appear in the table, or it is a status nothing can leave.
	assert.Len(t, orderTransitions, len(orderStatuses))

	agreementStatuses := map[string]bool{
		string(models.AgreementStatusDraft):     true,
		string(models.AgreementStatusConfirmed): true,
		string(models.AgreementStatusClosed):    true,
		string(models.AgreementStatusCancelled): true,
	}
	for from, targets := range agreementTransitions {
		assert.True(t, agreementStatuses[from], "unknown agreement status %q", from)
		for _, to := range targets {
			assert.True(t, agreementStatuses[to], "unknown agreement status %q", to)
		}
	}
	assert.Len(t, agreementTransitions, len(agreementStatuses))
}

// A refused transition must say something a user can act on; the two most likely mistakes get their
// own message.
func TestAssertOrderTransitionNamesTheProblem(t *testing.T) {
	testCases := []struct {
		name     string
		from, to models.PurchaseOrderStatus
		wantKey  string
	}{
		{
			"reviving a cancelled order", models.PurchaseOrderStatusCancelled,
			models.PurchaseOrderStatusPurchaseOrder, "purchase_order.cancelled_is_final",
		},
		{
			"re-confirming a committed order", models.PurchaseOrderStatusPurchaseOrder,
			models.PurchaseOrderStatusToApprove, "purchase_order.already_confirmed",
		},
		{
			"anything else", models.PurchaseOrderStatusToApprove,
			models.PurchaseOrderStatusRfq, "purchase_order.invalid_transition",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := ft.NewClientErrors()

			AssertOrderTransition(string(testCase.from), string(testCase.to), vErrs)

			assert.Equal(t, 1, vErrs.Count())
			// The key is what a client branches on; the message is only for a human reader.
			assert.Equal(t, testCase.wantKey, (*vErrs)[0].Key)
		})
	}
}

// A legal transition appends nothing.
func TestAssertOrderTransitionIsSilentWhenLegal(t *testing.T) {
	vErrs := ft.NewClientErrors()

	AssertOrderTransition(
		string(models.PurchaseOrderStatusRfq), string(models.PurchaseOrderStatusPurchaseOrder), vErrs)

	assert.Equal(t, 0, vErrs.Count())
}

// The approval decision, which routes a confirm to one status or the other.
func TestRequiresApproval(t *testing.T) {
	threshold := func(value string) *decimal.Decimal {
		parsed := dec(value)
		return &parsed
	}

	testCases := []struct {
		name      string
		mode      models.ApprovalMode
		threshold *decimal.Decimal
		total     string
		want      bool
	}{
		{
			// A threshold left over from a previous policy must not keep gating orders after the
			// organization has switched approvals off.
			name: "one step never requires approval, whatever the threshold",
			mode: models.ApprovalModeOneStep, threshold: threshold("1"), total: "1000000", want: false,
		},
		{
			name: "two step with no threshold approves everything",
			mode: models.ApprovalModeTwoStep, threshold: nil, total: "0.01", want: true,
		},
		{
			name: "two step above the threshold",
			mode: models.ApprovalModeTwoStep, threshold: threshold("1000"), total: "1000.01", want: true,
		},
		{
			// The comparison is "at or above", not "above".
			name: "two step exactly at the threshold",
			mode: models.ApprovalModeTwoStep, threshold: threshold("1000"), total: "1000", want: true,
		},
		{
			name: "two step below the threshold",
			mode: models.ApprovalModeTwoStep, threshold: threshold("1000"), total: "999.99", want: false,
		},
		{
			name: "a zero threshold means approve everything",
			mode: models.ApprovalModeTwoStep, threshold: threshold("0"), total: "0", want: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config := PurchaseConfiguration{
				ApprovalMode:      testCase.mode,
				ApprovalThreshold: testCase.threshold,
			}

			assert.Equal(t, testCase.want, RequiresApproval(config, dec(testCase.total)))
		})
	}
}

// An organization that never configured Purchase must still be able to buy things; defaulting to
// two-step would mean installing the module silently blocked every purchase.
func TestTheDefaultConfigurationBlocksNothing(t *testing.T) {
	config := DefaultPurchaseConfiguration()

	assert.Equal(t, models.ApprovalModeOneStep, config.ApprovalMode)
	assert.Equal(t, models.PoModificationPolicyAllowEdit, config.PoModificationPolicy)
	assert.Nil(t, config.ApprovalThreshold)
	assert.False(t, RequiresApproval(config, dec("999999999")))
}

func TestIsOrderOpenAndCommitted(t *testing.T) {
	assert.True(t, IsOrderOpen(string(models.PurchaseOrderStatusRfq)))
	assert.True(t, IsOrderOpen(string(models.PurchaseOrderStatusPurchaseOrder)))
	assert.False(t, IsOrderOpen(string(models.PurchaseOrderStatusCancelled)))

	assert.True(t, IsOrderCommitted(string(models.PurchaseOrderStatusPurchaseOrder)))
	assert.False(t, IsOrderCommitted(string(models.PurchaseOrderStatusToApprove)))
	assert.False(t, IsOrderCommitted(string(models.PurchaseOrderStatusRfq)))
}
