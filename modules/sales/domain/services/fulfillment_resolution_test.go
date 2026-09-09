package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The parts of fulfillment resolution that decide something without reading a row: the target
// selection strategy, the identity snapshot, and the reservation deadline.

func methodRecord(selection models.InitialTargetSelection) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesFulfillmentMethodFieldId:                     "FM1",
		models.SalesFulfillmentMethodFieldFulfillmentType:        string(models.FulfillmentTypeKioskDispense),
		models.SalesFulfillmentMethodFieldInitialTargetSelection: string(selection),
	}
}

// A method that delivers from where the sale was made resolves its own target. Nobody has to choose,
// because the customer is standing at the machine.
func TestCurrentOutletSelectionResolvesToTheOrdersPoint(t *testing.T) {
	request := FulfillmentMethodRequest{SalesPointId: "SP1"}

	targetId, refusal, err := resolveTargetOutletId(
		request, methodRecord(models.InitialTargetSelectionCurrentSalesOutlet))
	if err != nil {
		t.Fatalf("resolving a target must not need a repository: %v", err)
	}
	if refusal != nil {
		t.Fatalf("the order names a sales point, so the target is known: %v", refusal.ClientErrors)
	}
	if targetId != "SP1" {
		t.Errorf("the target must be the point the order was raised at, got %q", targetId)
	}
}

// THE refusal of target selection. A customer-chosen pickup with no choice made must be refused, not
// defaulted: guessing which kiosk somebody meant to collect from sends their goods to another town,
// and a refusal is recoverable where a wrong reservation is not.
func TestCustomerSelectedOutletRefusesWhenNoTargetWasNamed(t *testing.T) {
	request := FulfillmentMethodRequest{SalesPointId: "SP1"}

	targetId, refusal, err := resolveTargetOutletId(
		request, methodRecord(models.InitialTargetSelectionCustomerSelectedOutlet))
	if err != nil {
		t.Fatalf("resolving a target must not need a repository: %v", err)
	}
	if refusal == nil {
		t.Fatal("a customer-selected pickup with no choice made must be refused, not defaulted")
	}
	if targetId != "" {
		t.Errorf("a refused resolution must name no target, got %q", targetId)
	}
	if !hasReason(&refusal.ClientErrors, ReasonTargetRequired) {
		t.Errorf("the refusal must say a target is required, got %v", refusal.ClientErrors)
	}
}

// An explicit choice wins over the strategy: the client already answered the question the strategy
// exists to answer.
func TestAnExplicitTargetOverridesTheSelectionStrategy(t *testing.T) {
	request := FulfillmentMethodRequest{SalesPointId: "SP1", TargetOutletId: "SP2"}

	targetId, refusal, err := resolveTargetOutletId(
		request, methodRecord(models.InitialTargetSelectionCustomerSelectedOutlet))
	if err != nil || refusal != nil {
		t.Fatalf("a named target satisfies a customer-selected method: %v %v", err, refusal)
	}
	if targetId != "SP2" {
		t.Errorf("the client's choice must win, got %q", targetId)
	}
}

// A method with no strategy at all resolves nothing rather than falling back to the order's point.
// The absent value is a configuration mistake, and delivering somewhere by default would hide it.
func TestAnUnknownSelectionStrategyRefuses(t *testing.T) {
	_, refusal, err := resolveTargetOutletId(
		FulfillmentMethodRequest{SalesPointId: "SP1"}, dmodel.DynamicFields{})
	if err != nil {
		t.Fatalf("resolving a target must not need a repository: %v", err)
	}
	if refusal == nil {
		t.Error("a method declaring no selection strategy must refuse rather than default")
	}
}

// THE test of the identity snapshot. Absence of a principal must read as anonymous: a method that
// requires an identified buyer has to refuse rather than assume one, and failing open here would
// hand walk-up sales the policies written for people who can be contacted afterwards.
func TestAnUnauthenticatedContextIsAnonymous(t *testing.T) {
	if mode := DeriveCustomerIdentityMode(nil); mode != models.CustomerIdentityModeAnonymous {
		t.Errorf("no context at all must be anonymous, got %q", mode)
	}

	ctx := corectx.NewRequestContext(t.Context())
	if mode := DeriveCustomerIdentityMode(ctx); mode != models.CustomerIdentityModeAnonymous {
		t.Errorf("a context with no principal must be anonymous, got %q", mode)
	}
}

// A service principal is authenticated as a WORKLOAD, not as a buyer. A kiosk holding credentials
// says nothing about who is standing in front of it, and reading it as an identified customer would
// let every machine sale claim policies that need somebody to come back to.
func TestAServicePrincipalIsNotAnIdentifiedCustomer(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{
			Kind: corectx.PrincipalKindService,
			Id:   "KIOSK1",
		},
	})

	if mode := DeriveCustomerIdentityMode(ctx); mode != models.CustomerIdentityModeAnonymous {
		t.Errorf("a device credential does not identify a buyer, got %q", mode)
	}
}

func TestAUserPrincipalIsAnIdentifiedCustomer(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{
			Kind: corectx.PrincipalKindUser,
			Id:   "U1",
		},
	})

	if mode := DeriveCustomerIdentityMode(ctx); mode != models.CustomerIdentityModeAuthenticated {
		t.Errorf("a verified person is an identified customer, got %q", mode)
	}
}

// A null TTL must produce no deadline. Storing one would expire a sale that was dispensed seconds
// after payment and was already complete.
func TestNoTtlMeansNoReservationDeadline(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	if deadline := reservationDeadline(nil, now); deadline != nil {
		t.Errorf("a method with no TTL must not set a deadline, got %v", deadline)
	}

	zero := int32(0)
	if deadline := reservationDeadline(&zero, now); deadline != nil {
		t.Errorf("a non-positive TTL must not set a deadline, got %v", deadline)
	}
}

func TestATtlBecomesAnAbsoluteDeadline(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	ttl := int32(30)

	deadline := reservationDeadline(&ttl, now)
	if deadline == nil {
		t.Fatal("a positive TTL must produce a deadline")
	}
	if got := deadline.GoTime().UTC(); !got.Equal(now.Add(30 * time.Minute)) {
		t.Errorf("the deadline must be the TTL after now, got %v", got)
	}
}
