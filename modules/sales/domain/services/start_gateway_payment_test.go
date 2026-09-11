package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// What a gateway collection is for, and what a retry gets back. The money is the bill's answer, not
// the caller's: a client that believes it owes less must not be able to make that true by saying so.

// A caller that names no amount is collecting the whole of what is still owed, which is the
// ordinary case at a kiosk: the customer pays the bill, not a number the device chose.
func TestAnAmountLeftOutMeansEverythingStillOwed(t *testing.T) {
	amount, vErrs := resolvePayableAmount(dec("50000"), decimal.Zero)
	if vErrs != nil {
		t.Fatalf("resolving the whole outstanding balance was refused: %v", vErrs)
	}
	if !amount.Equal(dec("50000")) {
		t.Errorf("amount = %s, want 50000", amount)
	}
}

// Part of a bill may be collected: a customer splitting payment across two cards pays one of them
// now.
func TestAnAmountWithinTheBalanceIsTakenAsAsked(t *testing.T) {
	amount, vErrs := resolvePayableAmount(dec("50000"), dec("20000"))
	if vErrs != nil {
		t.Fatalf("a partial collection was refused: %v", vErrs)
	}
	if !amount.Equal(dec("20000")) {
		t.Errorf("amount = %s, want 20000", amount)
	}
}

// Overpayment is refused even where the cash policy would allow it. A customer handing over a note
// and taking change is one thing; a QR code asking for more than is owed is another, and the excess
// would have to go back through the provider rather than out of the till.
func TestAskingForMoreThanIsOwedIsRefused(t *testing.T) {
	_, vErrs := resolvePayableAmount(dec("50000"), dec("50001"))
	if vErrs == nil {
		t.Fatal("collecting more than the bill's outstanding balance must be refused")
	}
	if !hasViolation(vErrs, ReasonAmountExceedsDue) {
		t.Errorf("refusal = %v, want %s", vErrs, ReasonAmountExceedsDue)
	}
}

// A bill with nothing left to pay has no collection to start: showing a customer a QR code for a
// settled bill would take money the sale does not owe.
func TestABillWithNothingOutstandingCannotStartACollection(t *testing.T) {
	_, vErrs := resolvePayableAmount(decimal.Zero, decimal.Zero)
	if vErrs == nil {
		t.Fatal("a fully paid bill must not start a collection")
	}
	if !hasViolation(vErrs, ReasonNothingOutstanding) {
		t.Errorf("refusal = %v, want %s", vErrs, ReasonNothingOutstanding)
	}
}

// A retry with the same key is answered from the stored row, instructions included. The customer
// may already be looking at that QR code, and a second one for the same debt is the confusion the
// key exists to prevent.
func TestARetryIsAnsweredWithTheOriginalInstructions(t *testing.T) {
	existing := dmodel.DynamicFields{
		models.SalesPaymentFieldId:                "PAY1",
		models.SalesPaymentFieldSalesBillId:       "BILL1",
		models.SalesPaymentFieldPaymentOrderId:    "PTX-88281",
		models.SalesPaymentFieldProviderReference: "ORDER-CODE",
		models.SalesPaymentFieldQrCodeUrl:         "https://gw.example/qr/1",
		models.SalesPaymentFieldPayUrl:            "https://gw.example/pay/1",
	}

	replay := replayGatewayPayment(existing)

	if !replay.AlreadyStarted {
		t.Error("a replay must say so, or a caller cannot tell a reused QR from a new one")
	}
	if replay.SalesPaymentId != "PAY1" {
		t.Errorf("payment id = %q, want PAY1", replay.SalesPaymentId)
	}
	if replay.QrCodeUrl != "https://gw.example/qr/1" {
		t.Errorf("qr code = %q, want the stored one", replay.QrCodeUrl)
	}
	if replay.PaymentOrderId != "PTX-88281" {
		t.Errorf("payment order = %q, want the original PTX-88281", replay.PaymentOrderId)
	}
}

func hasViolation(vErrs *ft.ClientErrors, reason string) bool {
	if vErrs == nil {
		return false
	}
	for _, violation := range *vErrs {
		if violation.Key == reason {
			return true
		}
	}
	return false
}
