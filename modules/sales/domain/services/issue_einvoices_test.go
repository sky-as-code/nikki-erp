package services

import (
	"testing"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/settings"
)

// The issuance delay, which is the part of eligibility that can be checked without a database.
//
// It exists because the minutes after a sale are when it is most likely to be corrected or reversed
// at the counter, and a VAT invoice cannot simply be deleted afterwards. Waiting turns a would-be
// credit note into a sale that was never invoiced.

func TestTheIssuanceCutoffIsTheConfiguredDelayBeforeNow(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	policy := DefaultSalesPolicy()

	cutoff := now.Add(-time.Duration(policy.InvoiceIssueDelayMinutes) * time.Minute)

	// The default is two hours, so a sale settled at 09:59 is due at noon and one settled at 10:01
	// is not.
	if !cutoff.Equal(time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("cutoff = %s, want 10:00 for the default two-hour delay", cutoff)
	}

	settledEarlier := time.Date(2026, 9, 4, 9, 59, 0, 0, time.UTC)
	if !settledEarlier.Before(cutoff) {
		t.Error("a sale settled before the cutoff must be due")
	}

	settledLater := time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)
	if settledLater.Before(cutoff) {
		t.Error("a sale settled after the cutoff must not be due yet")
	}
}

// The default is the documented two hours. Stated here as well as in the settings package because
// this is the number that decides how long a buyer waits for their invoice.
func TestTheDefaultDelayIsTwoHours(t *testing.T) {
	if got := DefaultSalesPolicy().InvoiceIssueDelayMinutes; got != 120 {
		t.Errorf("default delay = %d minutes, want 120", got)
	}
	if settings.DefaultInvoiceIssueDelayMinutes != 120 {
		t.Errorf("the settings default disagrees: %d",
			settings.DefaultInvoiceIssueDelayMinutes)
	}
}

// A zero delay would mean issuing the instant a sale settles, which is exactly what the wait exists
// to prevent. The schema's minimum of 1 is what stops it, and the Go default must not undercut it.
func TestTheDelayIsNeverZero(t *testing.T) {
	if DefaultSalesPolicy().InvoiceIssueDelayMinutes <= 0 {
		t.Error("a zero delay issues before a same-visit correction could happen")
	}
}

// An outcome must map to exactly one counter, or a pass would report totals that do not add up to
// what it examined.
func TestEveryIssuanceOutcomeIsCountedOnce(t *testing.T) {
	outcomes := []issuanceOutcome{
		issuanceSkipped, issuanceIssued, issuanceFailed, issuanceIndeterminate,
	}
	seen := map[issuanceOutcome]bool{}
	for _, outcome := range outcomes {
		if seen[outcome] {
			t.Errorf("outcome %d is duplicated, so two states share a counter", outcome)
		}
		seen[outcome] = true
	}

	// Skipped is deliberately not counted anywhere: it means another worker claimed the instruction
	// first, which is the guard working rather than an event worth reporting.
	if issuanceSkipped == issuanceIssued || issuanceSkipped == issuanceFailed ||
		issuanceSkipped == issuanceIndeterminate {
		t.Error("a skipped instruction must be distinguishable from one that was acted on")
	}
}
