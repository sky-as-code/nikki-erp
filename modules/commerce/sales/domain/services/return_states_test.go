package services

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Acceptance tests for the two return decisions, plus the edges around them. Each names the rule
// it protects, because a failure here is a business error rather than a coding one.

// The headline case: goods back, money back, tax paperwork failed, and the return is nonetheless
// complete.
func TestReturnCompletesDespiteFailedFiscalAdjustment(t *testing.T) {
	status := DeriveReturnStatus(
		string(models.SalesReturnStatusProcessing),
		string(models.SalesReturnStepCompleted),
		string(models.SalesReturnStepCompleted),
	)

	if status != string(models.SalesReturnStatusCompleted) {
		t.Errorf("a return whose goods are back and money refunded must be completed, got %q."+
			" The fiscal adjustment is deliberately not an input to this decision (D-17): if it"+
			" were, a failed tax correction would hold open a return the customer has already"+
			" been made whole for.", status)
	}
}

// The same rule stated structurally. It is enforced by the SIGNATURE, not a branch:
// DeriveReturnStatus takes the two customer-facing steps and nothing else, so what is worth
// pinning is the arity. Assigning the function to a typed variable fails to compile the moment
// somebody adds a fiscal parameter, which is the change that would reintroduce the coupling.
func TestReturnCompletionSignatureExcludesFiscalStatus(t *testing.T) {
	var derive func(current, inventoryStatus, refundStatus string) string = DeriveReturnStatus

	status := derive(
		string(models.SalesReturnStatusProcessing),
		string(models.SalesReturnStepCompleted),
		string(models.SalesReturnStepCompleted),
	)
	if status != string(models.SalesReturnStatusCompleted) {
		t.Errorf("got %q, want completed", status)
	}
}

func TestReturnStaysProcessingUntilBothCustomerStepsSettle(t *testing.T) {
	cases := []struct {
		name      string
		inventory string
		refund    string
		want      string
		why       string
	}{
		{
			name:      "refund still pending",
			inventory: string(models.SalesReturnStepCompleted),
			refund:    string(models.SalesReturnStepPending),
			want:      string(models.SalesReturnStatusProcessing),
			why:       "the customer has not been paid back yet",
		},
		{
			name:      "goods not back yet",
			inventory: string(models.SalesReturnStepPending),
			refund:    string(models.SalesReturnStepCompleted),
			want:      string(models.SalesReturnStatusProcessing),
			why:       "the goods have not returned yet",
		},
		{
			name:      "refund failed",
			inventory: string(models.SalesReturnStepCompleted),
			refund:    string(models.SalesReturnStepFailed),
			want:      string(models.SalesReturnStatusProcessing),
			why: "a failed refund is a job somebody must finish, not an outcome: the customer" +
				" is still owed their money, so the return is not complete",
		},
		{
			name:      "services only, nothing to send back",
			inventory: string(models.SalesReturnStepNotRequired),
			refund:    string(models.SalesReturnStepCompleted),
			want:      string(models.SalesReturnStatusCompleted),
			why: "not_required counts as settled (D-32) — otherwise a services return waits" +
				" forever for a movement nobody will make",
		},
		{
			name:      "nothing owed either way",
			inventory: string(models.SalesReturnStepNotRequired),
			refund:    string(models.SalesReturnStepNotRequired),
			want:      string(models.SalesReturnStatusCompleted),
			why:       "a return against an unpaid order of services owes nothing in either direction",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DeriveReturnStatus(string(models.SalesReturnStatusProcessing), tc.inventory, tc.refund)
			if got != tc.want {
				t.Errorf("got %q, want %q — %s", got, tc.want, tc.why)
			}
		})
	}
}

func TestCancelledReturnIsNeverRevived(t *testing.T) {
	status := DeriveReturnStatus(
		string(models.SalesReturnStatusCancelled),
		string(models.SalesReturnStepCompleted),
		string(models.SalesReturnStepCompleted),
	)

	if status != string(models.SalesReturnStatusCancelled) {
		t.Errorf("a cancelled return must stay cancelled, got %q. Steps can report completion"+
			" after a cancellation through a late callback, and letting that revive the return"+
			" would resurrect a document somebody deliberately called off.", status)
	}
}

func TestReturnTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		allowed  bool
		why      string
	}{
		{"draft", "approved", true, "the ordinary path"},
		{"draft", "cancelled", true, "nothing irreversible has happened yet"},
		{"approved", "processing", true, "the side effects begin"},
		{"approved", "cancelled", true, "still nothing irreversible"},
		{"processing", "completed", true, "both customer-facing steps settled"},
		{
			from: "processing", to: "cancelled", allowed: false,
			why: "goods may already be moving or money already leaving; there is no honest way" +
				" to un-ask for either",
		},
		{"processing", "approved", false, "no route backwards once committed"},
		{"completed", "cancelled", false, "completed is terminal"},
		{"cancelled", "approved", false, "cancelled is terminal"},
		{"completed", "completed", true, "an idempotent retry of a status already held"},
	}

	for _, tc := range cases {
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			if got := CanTransitionReturn(tc.from, tc.to); got != tc.allowed {
				t.Errorf("CanTransitionReturn(%q, %q) = %v, want %v — %s",
					tc.from, tc.to, got, tc.allowed, tc.why)
			}
		})
	}
}

// The acceptance tests for the returnable-quantity basis.
func TestReturnableQuantity(t *testing.T) {
	cases := []struct {
		name string
		line ReturnableLine
		want string
		why  string
	}{
		{
			name: "physical product, partially fulfilled and partially returned",
			line: ReturnableLine{
				Ordered: dec("5"), Fulfilled: dec("4"), PreviouslyReturned: dec("1"),
				RequiresFulfillment: true,
			},
			want: "3",
			why:  "delivered minus already returned; the fifth was never handed over",
		},
		{
			name: "service, never fulfilled",
			line: ReturnableLine{
				Ordered: dec("5"), Fulfilled: dec("0"), PreviouslyReturned: dec("1"),
				RequiresFulfillment: false,
			},
			want: "4",
			why: "ordered minus already returned. Using fulfilled here would give zero and make" +
				" refunding a service impossible, which BR 52 explicitly requires to work",
		},
		{
			name: "the answer's worked example for goods",
			line: ReturnableLine{
				Ordered: dec("10"), Fulfilled: dec("7"), PreviouslyReturned: dec("2"),
				RequiresFulfillment: true,
			},
			want: "5",
			why:  "7 handed over, 2 already back",
		},
		{
			name: "the answer's worked example for services",
			line: ReturnableLine{
				Ordered: dec("3"), Fulfilled: dec("0"), PreviouslyReturned: dec("1"),
				RequiresFulfillment: false,
			},
			want: "2",
			why:  "3 ordered, 1 already returned, fulfilment irrelevant",
		},
		{
			name: "fully returned goods",
			line: ReturnableLine{
				Ordered: dec("4"), Fulfilled: dec("4"), PreviouslyReturned: dec("4"),
				RequiresFulfillment: true,
			},
			want: "0",
			why:  "nothing left to send back",
		},
		{
			name: "ordered but nothing delivered",
			line: ReturnableLine{
				Ordered: dec("6"), Fulfilled: dec("0"), PreviouslyReturned: dec("0"),
				RequiresFulfillment: true,
			},
			want: "0",
			why: "a customer cannot return goods they never received — this is exactly the case" +
				" the services exception above must NOT be allowed to leak into",
		},
		{
			name: "over-returned through data repair",
			line: ReturnableLine{
				Ordered: dec("2"), Fulfilled: dec("2"), PreviouslyReturned: dec("3"),
				RequiresFulfillment: true,
			},
			want: "0",
			why: "never negative: a negative allowance reads as permission to return minus one" +
				" and underflows any comparison made against it",
		},
		{
			name: "fractional quantity",
			line: ReturnableLine{
				Ordered: dec("2.5"), Fulfilled: dec("2.5"), PreviouslyReturned: dec("0.75"),
				RequiresFulfillment: true,
			},
			want: "1.75",
			why:  "quantities carry six decimal places; the arithmetic must not round",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReturnableQuantity(tc.line)
			if !got.Equal(dec(tc.want)) {
				t.Errorf("ReturnableQuantity = %s, want %s — %s", got, tc.want, tc.why)
			}
		})
	}
}

// The rule that makes the two bases genuinely different. If this passes with both bases equal,
// the exception has been lost.
func TestServiceAndGoodsDisagreeWhenNothingWasDelivered(t *testing.T) {
	ordered, fulfilled, returned := dec("5"), dec("0"), dec("0")

	goods := ReturnableQuantity(ReturnableLine{
		Ordered: ordered, Fulfilled: fulfilled, PreviouslyReturned: returned,
		RequiresFulfillment: true,
	})
	service := ReturnableQuantity(ReturnableLine{
		Ordered: ordered, Fulfilled: fulfilled, PreviouslyReturned: returned,
		RequiresFulfillment: false,
	})

	if !goods.IsZero() {
		t.Errorf("undelivered goods must not be returnable, got %s", goods)
	}
	if !service.Equal(dec("5")) {
		t.Errorf("an undelivered service must be returnable against what was ordered, got %s", service)
	}
}

func TestDeriveInventoryStepStatus(t *testing.T) {
	goods := ReturnableLine{RequiresFulfillment: true}
	service := ReturnableLine{RequiresFulfillment: false}

	cases := []struct {
		name  string
		lines []ReturnableLine
		want  string
		why   string
	}{
		{
			name: "services only", lines: []ReturnableLine{service, service},
			want: string(models.SalesReturnStepNotRequired),
			why:  "no goods owed, so the return completes on the refund alone",
		},
		{
			name: "goods only", lines: []ReturnableLine{goods},
			want: string(models.SalesReturnStepPending),
			why:  "goods must physically come back",
		},
		{
			name: "mixed", lines: []ReturnableLine{service, goods},
			want: string(models.SalesReturnStepPending),
			why: "a part-physical return still needs its physical half; treating it as" +
				" not_required would complete the return with goods still at the customer",
		},
		{
			name: "no lines at all", lines: nil,
			want: string(models.SalesReturnStepNotRequired),
			why:  "nothing to send back",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeriveInventoryStepStatus(tc.lines); got != tc.want {
				t.Errorf("got %q, want %q — %s", got, tc.want, tc.why)
			}
		})
	}
}

func TestDeriveRefundStepStatus(t *testing.T) {
	if got := DeriveRefundStepStatus(decimal.Zero); got != string(models.SalesReturnStepNotRequired) {
		t.Errorf("a return owing no money must not wait on a refund, got %q — this covers a"+
			" return against an order that was never paid", got)
	}
	if got := DeriveRefundStepStatus(dec("48000")); got != string(models.SalesReturnStepPending) {
		t.Errorf("money owed must start pending, got %q", got)
	}
}
