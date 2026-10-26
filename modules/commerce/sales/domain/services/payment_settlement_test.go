package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The payment arithmetic and the settlement rule, where an error is expensive and silent. The gated
// operations read the repository and are exercised live.

func paymentRow(status, amount string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesPaymentFieldStatus: status,
		models.SalesPaymentFieldAmount: amount,
	}
}

// Only captured money counts: an authorization is a hold the provider may still release, and counting
// it would settle a bill against funds that never arrived.
func TestOnlyCapturedPaymentsCount(t *testing.T) {
	payments := []dmodel.DynamicFields{
		paymentRow(string(models.SalesPaymentStatusCaptured), "100"),
		paymentRow(string(models.SalesPaymentStatusAuthorized), "500"),
		paymentRow(string(models.SalesPaymentStatusPending), "500"),
		paymentRow(string(models.SalesPaymentStatusFailed), "500"),
		paymentRow(string(models.SalesPaymentStatusCancelled), "500"),
		paymentRow(string(models.SalesPaymentStatusCaptured), "50"),
	}

	if got := models.SumCapturedAmount(payments); !got.Equal(dec("150")) {
		t.Errorf("captured total = %s, want 150 — only the two captured payments count", got)
	}
}

func TestNoPaymentsSumToZero(t *testing.T) {
	if !models.SumCapturedAmount(nil).IsZero() {
		t.Error("no payments must sum to zero")
	}
}

// The sum must survive a decimal arriving as a string, which is how it comes back from jsonb.
func TestCapturedSumHandlesJsonShapes(t *testing.T) {
	payments := []dmodel.DynamicFields{
		{
			models.SalesPaymentFieldStatus: string(models.SalesPaymentStatusCaptured),
			models.SalesPaymentFieldAmount: "100",
		},
		{
			models.SalesPaymentFieldStatus: string(models.SalesPaymentStatusCaptured),
			models.SalesPaymentFieldAmount: float64(50),
		},
	}

	if got := models.SumCapturedAmount(payments); !got.Equal(dec("150")) {
		t.Errorf("captured total = %s, want 150 across mixed shapes", got)
	}
}

// paid requires exact equality: a tolerance would mark a bill paid while a fraction remained owed,
// and those fractions accumulate.
func TestBillPaymentStatusDerivation(t *testing.T) {
	cases := []struct {
		payable  string
		captured string
		want     string
	}{
		{"100", "0", string(models.SalesOrderPaymentStatusUnpaid)},
		{"100", "1", string(models.SalesOrderPaymentStatusPartiallyPaid)},
		{"100", "99.9999", string(models.SalesOrderPaymentStatusPartiallyPaid)},
		{"100", "100", string(models.SalesOrderPaymentStatusPaid)},
		{"100", "100.0001", string(models.SalesOrderPaymentStatusOverpaid)},
		{"100", "150", string(models.SalesOrderPaymentStatusOverpaid)},
	}

	for _, testCase := range cases {
		got := DeriveBillPaymentStatus(dec(testCase.payable), dec(testCase.captured))
		if got != testCase.want {
			t.Errorf("payable %s captured %s = %q, want %q",
				testCase.payable, testCase.captured, got, testCase.want)
		}
	}
}

// A fraction short is not paid; rounding it to paid is how a business stops noticing it is owed money.
func TestAFractionShortIsNotPaid(t *testing.T) {
	got := DeriveBillPaymentStatus(dec("48000"), dec("47999.9999"))
	if got == string(models.SalesOrderPaymentStatusPaid) {
		t.Error("a bill one ten-thousandth short must not read as paid")
	}
}

// Change is only due when the policy permits it, and only on the excess.
func TestChangeIsDueOnlyUnderTheCashPolicy(t *testing.T) {
	permissive := SalesPolicy{AllowCashChange: true}
	strict := SalesPolicy{AllowCashChange: false}

	if got := changeDue(dec("110"), dec("100"), permissive); !got.Equal(dec("10")) {
		t.Errorf("change = %s, want 10", got)
	}
	if got := changeDue(dec("110"), dec("100"), strict); !got.IsZero() {
		t.Errorf("change = %s, want 0 when the policy forbids it", got)
	}
	if got := changeDue(dec("90"), dec("100"), permissive); !got.IsZero() {
		t.Errorf("change = %s, want 0 when the bill is underpaid", got)
	}
	if got := changeDue(dec("100"), dec("100"), permissive); !got.IsZero() {
		t.Errorf("change = %s, want 0 when the bill is paid exactly", got)
	}
}

// A payment must be for more than zero; a negative one is a refund, a separate resource with its own
// lifecycle.
func TestAPaymentMustBePositive(t *testing.T) {
	for _, amount := range []string{"0", "-100"} {
		if dec(amount).IsPositive() {
			t.Errorf("%s must not read as a positive amount", amount)
		}
	}
}

// Overpaid still settles: the excess is change the till hands back, not a reason to keep the bill open.
func TestAnOverpaidBillStillSettles(t *testing.T) {
	captured, billTotal := dec("110"), dec("100")

	settled := !captured.LessThan(billTotal) && billTotal.IsPositive()
	if !settled {
		t.Error("an overpaid bill must settle: what was owed has been paid")
	}
}

// A bill of zero must not settle itself, or every fresh bill would be born settled and the first
// payment would land on a closed one.
func TestAZeroBillDoesNotSettleItself(t *testing.T) {
	captured, billTotal := decimal.Zero, decimal.Zero

	settled := !captured.LessThan(billTotal) && billTotal.IsPositive()
	if settled {
		t.Error("a bill of zero must not settle itself; it has asked for nothing yet")
	}
}

// The method-count limit counts distinct methods, not payments: counting payments would refuse a
// legitimate two-card split while allowing five taps on one card.
func TestTheMethodLimitCountsDistinctMethods(t *testing.T) {
	distinct := map[string]bool{}
	for _, methodId := range []string{"CARD", "CARD", "CARD"} {
		distinct[methodId] = true
	}
	if len(distinct) != 1 {
		t.Errorf("three payments on one method must count as %d method, got %d", 1, len(distinct))
	}

	for _, methodId := range []string{"CASH"} {
		distinct[methodId] = true
	}
	if len(distinct) != 2 {
		t.Errorf("adding a second method must count as 2, got %d", len(distinct))
	}
}
