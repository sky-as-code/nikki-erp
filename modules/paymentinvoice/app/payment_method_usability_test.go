package app

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
	it "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

// These tests pin the three usability gates, which are the whole reason this port exists: none of
// them is readable off the row, and a consumer re-deriving them would be silently wrong.
//
// They need no database. judge takes a record and answers, so the gates are testable directly,
// which is what makes it worth having them in one function rather than spread through the two
// public methods.

// stubGateway is the smallest thing the registry accepts.
type stubGateway struct {
	itGateway.PaymentGateway
	code string
}

func (this *stubGateway) AdapterCode() string { return this.code }

// serviceWithAdapters builds the service with a registry holding exactly the named adapters, which
// is how a deployment's build is simulated.
func serviceWithAdapters(t *testing.T, codes ...string) *PaymentMethodApplicationServiceImpl {
	t.Helper()
	registry := itGateway.NewRegistry()
	for _, code := range codes {
		if err := registry.Register(&stubGateway{code: code}); err != nil {
			t.Fatalf("registering stub adapter %q: %v", code, err)
		}
	}
	return &PaymentMethodApplicationServiceImpl{registry: registry}
}

func methodRecord(overrides dmodel.DynamicFields) dmodel.DynamicFields {
	record := dmodel.DynamicFields{
		models.PaymentMethodFieldId:          "01M3PM00000000000000000001",
		models.PaymentMethodFieldCode:        "momo",
		models.PaymentMethodFieldAdapterCode: "momo",
		models.PaymentMethodFieldIsActive:    true,
		basemodel.FieldIsArchived:            false,
	}
	for key, value := range overrides {
		record[key] = value
	}
	return record
}

func TestUsableWhenActiveAndAdapterShips(t *testing.T) {
	service := serviceWithAdapters(t, "momo")

	data := service.judge(methodRecord(nil), nil)
	if !data.IsUsable {
		t.Errorf("a method that is active with its adapter present must be usable, got reason %q",
			data.UnusableReason)
	}
}

// A nil is_active counts as inactive, mirroring order_records.go:76. The alternative — treating a
// missing flag as a default of true — would let a half-written row take money.
func TestNilIsActiveCountsAsInactive(t *testing.T) {
	service := serviceWithAdapters(t, "momo")

	data := service.judge(methodRecord(dmodel.DynamicFields{
		models.PaymentMethodFieldIsActive: nil,
	}), nil)
	if data.IsUsable {
		t.Error("a nil is_active must count as inactive, never as a default of true")
	}
	if data.UnusableReason != it.ReasonInactive {
		t.Errorf("reason = %q, want %q", data.UnusableReason, it.ReasonInactive)
	}
}

// The gate that cannot be derived from the row at all: the same data is usable on a build shipping
// the adapter and unusable on one that does not.
func TestAdapterGateIsDeploymentDependent(t *testing.T) {
	record := methodRecord(dmodel.DynamicFields{
		models.PaymentMethodFieldAdapterCode: "mbbank",
	})

	withAdapter := serviceWithAdapters(t, "mbbank").judge(record, nil)
	if !withAdapter.IsUsable {
		t.Errorf("a build shipping the mbbank adapter must report it usable, got %q",
			withAdapter.UnusableReason)
	}

	withoutAdapter := serviceWithAdapters(t, "momo").judge(record, nil)
	if withoutAdapter.IsUsable {
		t.Error("a build with no mbbank adapter must report the method unusable: " +
			"this is the gate a consumer cannot see from the row")
	}
	if withoutAdapter.UnusableReason != it.ReasonGatewayUnavailable {
		t.Errorf("reason = %q, want %q",
			withoutAdapter.UnusableReason, it.ReasonGatewayUnavailable)
	}
}

// The asymmetry that would be a bug if it were not deliberate: min is inclusive, max is EXCLUSIVE
// (order_domservice.go:341). A port answering "usable" for an amount CreatePayment then refused
// would be worse than no port at all, so it is copied rather than corrected.
func TestAmountBoundsAreInclusiveMinExclusiveMax(t *testing.T) {
	service := serviceWithAdapters(t, "momo")
	record := methodRecord(dmodel.DynamicFields{
		models.PaymentMethodFieldMinAmount: decimal.NewFromInt(1000),
		models.PaymentMethodFieldMaxAmount: decimal.NewFromInt(20000),
	})

	cases := []struct {
		name       string
		amount     int64
		wantUsable bool
	}{
		{"below the minimum", 999, false},
		{"exactly the minimum is accepted", 1000, true},
		{"inside", 15000, true},
		{"one below the maximum", 19999, true},
		{"exactly the maximum is REFUSED", 20000, false},
		{"above", 20001, false},
		{"zero", 0, false},
		{"negative", -1, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			amount := decimal.NewFromInt(testCase.amount)
			data := service.judge(record, &amount)
			if data.IsUsable != testCase.wantUsable {
				t.Errorf("amount %d: usable = %v, want %v (reason %q)",
					testCase.amount, data.IsUsable, testCase.wantUsable, data.UnusableReason)
			}
		})
	}
}

// With no amount the bounds are not applied at all: a listing decides whether a method may ever be
// offered, not whether one particular payment would pass.
func TestNilAmountSkipsTheBoundsGate(t *testing.T) {
	service := serviceWithAdapters(t, "momo")

	data := service.judge(methodRecord(dmodel.DynamicFields{
		models.PaymentMethodFieldMinAmount: decimal.NewFromInt(1000),
		models.PaymentMethodFieldMaxAmount: decimal.NewFromInt(20000),
	}), nil)
	if !data.IsUsable {
		t.Errorf("a method with bounds and no amount to check must be usable, got %q",
			data.UnusableReason)
	}
}

// Archived beats inactive beats gateway beats amount. The order matters: a caller shown "amount
// too large" for a method that is also archived would go and fix the wrong thing.
func TestTheMostPermanentObstacleIsReported(t *testing.T) {
	service := serviceWithAdapters(t) // no adapters at all
	amount := decimal.NewFromInt(1)

	archived := service.judge(methodRecord(dmodel.DynamicFields{
		basemodel.FieldIsArchived:         true,
		models.PaymentMethodFieldIsActive: false,
	}), &amount)
	if archived.UnusableReason != it.ReasonArchived {
		t.Errorf("archived must outrank inactive, got %q", archived.UnusableReason)
	}

	inactive := service.judge(methodRecord(dmodel.DynamicFields{
		models.PaymentMethodFieldIsActive: false,
	}), &amount)
	if inactive.UnusableReason != it.ReasonInactive {
		t.Errorf("inactive must outrank the missing gateway, got %q", inactive.UnusableReason)
	}
}

// A service built without a registry must refuse everything rather than treat "no registry" as
// "every adapter present". This is the fail-closed direction, and it is the one that matters:
// failing open would let a wiring bug authorize payments through adapters that do not exist.
func TestNoRegistryFailsClosed(t *testing.T) {
	service := &PaymentMethodApplicationServiceImpl{}

	data := service.judge(methodRecord(nil), nil)
	if data.IsUsable {
		t.Error("a service with no gateway registry must report every method unusable")
	}
	if data.UnusableReason != it.ReasonGatewayUnavailable {
		t.Errorf("reason = %q, want %q", data.UnusableReason, it.ReasonGatewayUnavailable)
	}
}
