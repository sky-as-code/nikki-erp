package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The invoicing rules that can be pinned without a repository. The gated operation reads bills and
// their allocations and is exercised live; what is pinned here is the part where being wrong means
// issuing a legal document that should not exist, or refusing one that should.

func fiscalRow(intent, status string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesFiscalRequestFieldIntent: intent,
		models.SalesFiscalRequestFieldStatus: status,
	}
}

// THE rule of BR 77. Only a provider-confirmed request is issued; everything else is not, and
// `pending` most of all — a request in flight has produced no document a customer can deduct.
func TestOnlyAConfirmedRequestReadsAsIssued(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{string(models.SalesFiscalStatusIssued), true},
		{string(models.SalesFiscalStatusPending), false},
		{string(models.SalesFiscalStatusFailed), false},
		{string(models.SalesFiscalStatusCancelled), false},
	}

	for _, testCase := range cases {
		record := fiscalRow(string(models.SalesFiscalIntentIssueOriginal), testCase.status)
		if got := models.NewSalesFiscalRequestFrom(record).IsIssued(); got != testCase.want {
			t.Errorf("status %q reads as issued=%v, want %v", testCase.status, got, testCase.want)
		}
	}
}

// BOTH pending and issued block a second original invoice, and the pending half is the one worth
// the test. Sales does not know whether an in-flight request became issued; assuming it did not is
// exactly how a sale acquires two VAT invoices, which is a tax filing to correct.
func TestPendingBlocksASecondOriginalInvoice(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{string(models.SalesFiscalStatusIssued), true},
		{string(models.SalesFiscalStatusPending), true},
		{string(models.SalesFiscalStatusFailed), false},
		{string(models.SalesFiscalStatusCancelled), false},
	}

	for _, testCase := range cases {
		record := fiscalRow(string(models.SalesFiscalIntentIssueOriginal), testCase.status)
		got := models.NewSalesFiscalRequestFrom(record).BlocksNewRequest()
		if got != testCase.want {
			t.Errorf("status %q blocks=%v, want %v — a pending request may already have issued "+
				"a document Sales has not heard about", testCase.status, got, testCase.want)
		}
	}
}

// A record with no status at all must not read as issued. Absent is not confirmed, and the default
// reading here has to be the safe one: the opposite would report a document nobody issued.
func TestAnAbsentStatusIsNotIssued(t *testing.T) {
	empty := models.NewSalesFiscalRequestFrom(dmodel.DynamicFields{})
	if empty.IsIssued() {
		t.Error("a request with no status must not read as issued")
	}
	if empty.IsInFlight() {
		t.Error("a request with no status must not read as in flight either; it is neither, and " +
			"guessing between them is what BlocksNewRequest exists to avoid guessing about")
	}
}

// The buyer gate checks the two fields that make an invoice valid, and NOT the two that merely make
// it useful. Requiring an address would refuse invoices a provider would have accepted, which is
// Sales deciding invoice law by the back door (BR 46).
func TestBuyerCompletenessChecksTaxCodeAndNameOnly(t *testing.T) {
	cases := []struct {
		name  string
		buyer itInvoicing.BuyerInfo
		want  string
	}{
		{
			name:  "both missing",
			buyer: itInvoicing.BuyerInfo{},
			want:  "a tax code and a legal name",
		},
		{
			name:  "tax code missing",
			buyer: itInvoicing.BuyerInfo{LegalName: "ACME Co"},
			want:  "a tax code",
		},
		{
			name:  "legal name missing",
			buyer: itInvoicing.BuyerInfo{TaxCode: "0101234567"},
			want:  "a legal name",
		},
		{
			name:  "complete without address or email",
			buyer: itInvoicing.BuyerInfo{TaxCode: "0101234567", LegalName: "ACME Co"},
			want:  "",
		},
	}

	for _, testCase := range cases {
		if got := missingBuyerFields(testCase.buyer); got != testCase.want {
			t.Errorf("%s: missing = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

// The four business intents are accepted and a document type is not. The negative half is the point:
// a caller that could pass "credit_note" would be choosing the legal document, which is the
// provider's decision (BR 50).
func TestOnlyBusinessIntentsAreAccepted(t *testing.T) {
	for _, intent := range []string{
		string(models.SalesFiscalIntentIssueOriginal),
		string(models.SalesFiscalIntentAdjustForFullReturn),
		string(models.SalesFiscalIntentAdjustForPartialReturn),
		string(models.SalesFiscalIntentAdjustPrice),
	} {
		if !isKnownFiscalIntent(intent) {
			t.Errorf("%q must be accepted", intent)
		}
	}

	for _, notAnIntent := range []string{
		"credit_note", "invoice", "vat_invoice", "adjust", "", "ISSUE",
	} {
		if isKnownFiscalIntent(notAnIntent) {
			t.Errorf("%q must be refused: it names a document or nothing at all, not what "+
				"commercially happened", notAnIntent)
		}
	}
}

// The Sales intent constants and the port's must agree exactly. They are declared separately - the
// model must not import the port, and the port must not import the model - so nothing but a test
// stops them drifting, and a drift would send the provider an intent it does not recognise.
func TestModelAndPortIntentsAgree(t *testing.T) {
	pairs := []struct {
		model models.SalesFiscalIntent
		port  itInvoicing.FiscalIntent
	}{
		{models.SalesFiscalIntentIssueOriginal, itInvoicing.IntentIssueOriginal},
		{models.SalesFiscalIntentAdjustForFullReturn, itInvoicing.IntentAdjustForFullReturn},
		{models.SalesFiscalIntentAdjustForPartialReturn, itInvoicing.IntentAdjustForPartialReturn},
		{models.SalesFiscalIntentAdjustPrice, itInvoicing.IntentAdjustPrice},
	}

	for _, pair := range pairs {
		if string(pair.model) != string(pair.port) {
			t.Errorf("intent mismatch: model %q vs port %q", pair.model, pair.port)
		}
	}
}

// The buyer snapshot stores what was supplied, under the names the column is read back by. Stored
// as a map rather than the struct so the read path does not depend on Go field names (BR 87.7).
func TestBuyerSnapshotFreezesWhatWasSupplied(t *testing.T) {
	snapshot := buyerSnapshotOf(itInvoicing.BuyerInfo{
		TaxCode:   "0101234567",
		LegalName: "ACME Co",
		Address:   "1 Nguyen Hue",
		Email:     "ke-toan@acme.example",
	})

	for field, want := range map[string]string{
		"tax_code":   "0101234567",
		"legal_name": "ACME Co",
		"address":    "1 Nguyen Hue",
		"email":      "ke-toan@acme.example",
	} {
		if got, _ := snapshot[field].(string); got != want {
			t.Errorf("snapshot[%q] = %q, want %q", field, got, want)
		}
	}
}

// A provider message longer than the column must not cost the whole record of why an invoice failed.
func TestALongProviderErrorIsTruncatedNotDropped(t *testing.T) {
	long := make([]byte, 4000)
	for index := range long {
		long[index] = 'x'
	}

	got := truncateError(string(long))
	if len(got) != 1000 {
		t.Errorf("truncated length = %d, want 1000", len(got))
	}
	if truncateError("short") != "short" {
		t.Error("a message that fits must be kept whole")
	}
}
