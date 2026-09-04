package services

import (
	"testing"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The gateway gate, tested without a database: it asks the payment-method port one question and
// decides on the answer, so a fake port is the whole fixture.

type fakeMethodPort struct {
	data    itExt.PaymentMethodData
	hasData bool
	err     error
}

func (this *fakeMethodPort) ListPaymentMethods(
	corectx.Context, itExt.ListPaymentMethodsQuery,
) (*itExt.ListPaymentMethodsResult, error) {
	return &itExt.ListPaymentMethodsResult{}, nil
}

func (this *fakeMethodPort) AssertUsable(
	corectx.Context, itExt.AssertUsableQuery,
) (*itExt.AssertUsableResult, error) {
	if this.err != nil {
		return nil, this.err
	}
	return &itExt.AssertUsableResult{Data: this.data, HasData: this.hasData}, nil
}

func gatewayParams() StartGatewayPaymentParams {
	return StartGatewayPaymentParams{
		SalesBillId:     "bill-1",
		PaymentMethodId: "method-1",
		Amount:          decimal.RequireFromString("10.00"),
	}
}

// Cash cannot be collected through a gateway. Sent down this path it would open an order nobody can
// pay, and the till would show a QR code for money already in the drawer.
func TestGatewayGateRefusesAMethodWithoutAGateway(t *testing.T) {
	port := &fakeMethodPort{
		hasData: true,
		data:    itExt.PaymentMethodData{Code: "cash", HasGateway: false},
	}

	vErrs, err := assertMethodCollectsThroughGateway(nil, gatewayParams(), port)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vErrs == nil || vErrs.Count() == 0 {
		t.Fatal("a method with no gateway must be refused")
	}
	if !vErrs.Has("payment_method_id") {
		t.Errorf("the refusal must name the offending field, got %v", *vErrs)
	}
}

func TestGatewayGateAcceptsAMethodWithAGateway(t *testing.T) {
	port := &fakeMethodPort{
		hasData: true,
		data:    itExt.PaymentMethodData{Code: "momo", HasGateway: true},
	}

	vErrs, err := assertMethodCollectsThroughGateway(nil, gatewayParams(), port)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vErrs != nil {
		t.Errorf("a gateway-backed method must pass the gate, got %v", *vErrs)
	}
}

// Default-deny, matching every other unavailable port in this module: an absent port refuses rather
// than waving the payment through to a gateway that is not there.
func TestGatewayGateRefusesWhenThePortIsMissing(t *testing.T) {
	vErrs, err := assertMethodCollectsThroughGateway(nil, gatewayParams(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vErrs == nil || vErrs.Count() == 0 {
		t.Fatal("a missing payment-method port must refuse")
	}
}

// A method the upstream will not vouch for is refused before an order is opened: learning at the
// gateway what Sales could have known locally costs a real round trip.
func TestGatewayGateRefusesAnUnusableMethod(t *testing.T) {
	port := &fakeMethodPort{hasData: false}

	vErrs, err := assertMethodCollectsThroughGateway(nil, gatewayParams(), port)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vErrs == nil || vErrs.Count() == 0 {
		t.Fatal("a method the upstream does not vouch for must be refused")
	}
}
