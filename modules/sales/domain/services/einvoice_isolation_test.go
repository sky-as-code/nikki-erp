package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CR §62: no fulfillment code path may reach an invoice workflow.
//
// A refund raised because a machine could not hand goods over is a fulfillment event, not a
// commercial revision of the sale: the customer bought and paid for goods they simply never
// received, and asking a tax authority to correct a document over a jammed slot would file an
// adjustment for a sale that happened exactly as invoiced.
//
// This is asserted structurally rather than only behaviourally, because the failure mode is a call
// somebody adds later in good faith. A behavioural test would only catch it if that path happened to
// be exercised; this catches the import itself.

// fulfillmentOwnedFiles are the files implementing the fulfillment feature. A new one added to the
// feature belongs in this list.
var fulfillmentOwnedFiles = []string{
	"create_fulfillment.go",
	"create_attempt.go",
	"apply_attempt_result.go",
	"fulfillment_failure_policy.go",
	"fulfillment_expiry.go",
	"fulfillment_refund.go",
	"fulfillment_rollup.go",
	"fulfillment_view.go",
	"reassign_target.go",
	"release_fulfillment.go",
	"resolve_fulfillment_method.go",
	"refund_settlement_sync.go",
}

// fiscalSymbols are the ways this package can start an invoice workflow.
var fiscalSymbols = []string{
	"RequestInvoice(",
	"IssueEInvoices(",
	"buildIssueRequest(",
	"invoicing.",
}

func TestNoFulfillmentPathTriggersAnInvoiceWorkflow(t *testing.T) {
	for _, name := range fulfillmentOwnedFiles {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Clean(name))
			if err != nil {
				t.Fatalf("the fulfillment file list is stale; %s could not be read: %v", name, err)
			}
			body := string(source)

			for _, symbol := range fiscalSymbols {
				if strings.Contains(body, symbol) {
					t.Errorf("%s calls %s — CR §62 forbids a fulfillment path from reaching an "+
						"invoice workflow; a jammed slot must not file a tax correction",
						name, strings.TrimSuffix(symbol, "("))
				}
			}
		})
	}
}

// The fiscal step of a return must skip a fulfillment-failure refund outright. This pins the guard
// itself, so removing it fails here rather than silently in production.
func TestTheFiscalStepGuardsAgainstFulfillmentFailureRefunds(t *testing.T) {
	source, err := os.ReadFile(filepath.Clean("process_return.go"))
	if err != nil {
		t.Fatalf("could not read process_return.go: %v", err)
	}
	body := string(source)

	if !strings.Contains(body, "IsFulfillmentFailure()") {
		t.Error("the fiscal adjustment step must skip fulfillment-failure refunds (CR §62); " +
			"the guard reading IsFulfillmentFailure has gone")
	}

	guardAt := strings.Index(body, "IsFulfillmentFailure()")
	callAt := strings.Index(body, "RequestInvoice(")
	if guardAt < 0 || callAt < 0 || guardAt > callAt {
		t.Error("the fulfillment-failure guard must come BEFORE the invoice request, or an " +
			"automatic refund would file a fiscal adjustment before anything checked")
	}
}
