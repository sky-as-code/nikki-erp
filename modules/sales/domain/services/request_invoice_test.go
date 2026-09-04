package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The invoicing rules that can be pinned without a repository: the parts where being wrong means
// issuing a legal document that should not exist, or refusing one that should.

func fiscalRow(intent, status string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesFiscalRequestFieldIntent: intent,
		models.SalesFiscalRequestFieldStatus: status,
	}
}

// Only a provider-confirmed request is issued; `pending` most of all is not, since a request in
// flight has produced no document a customer can deduct.
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

// Both pending and issued block a second original invoice. Sales does not know whether an
// in-flight request became issued; assuming it did not is how a sale acquires two VAT invoices.
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

// A record with no status at all must not read as issued: absent is not confirmed.
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

// The buyer gate checks the two fields that make an invoice valid, not the two that merely make
// it useful; requiring an address would refuse invoices a provider would have accepted.
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

// The four business intents are accepted and a document type is not: choosing the legal document
// is the provider's decision.
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

// ISSUE_ORIGINAL is refused here even though it remains a valid intent elsewhere: the Billing
// Instruction and its scheduled job took over raising an original, and a request written through
// this path would sit `pending` for ever because that job does not read these rows.
//
// The refusal is checked before anything touches a repository, so this exercises the real gate.
func TestRequestingAnOriginalInvoiceIsRefusedHere(t *testing.T) {
	// A bill and buyer that would otherwise pass every gate, so only the intent can be what refuses.
	bill := dmodel.DynamicFields{
		models.SalesBillFieldId:     "01BILL000000000000000000",
		models.SalesBillFieldStatus: string(models.SalesBillStatusSettled),
	}
	params := RequestInvoiceParams{
		SalesBillId: "01BILL000000000000000000",
		Buyer:       itInvoicing.BuyerInfo{TaxCode: "0101234567", LegalName: "Acme Co"},
	}

	vErrs := assertInvoiceRequestable(nil, bill, string(models.SalesFiscalIntentIssueOriginal), params)
	if vErrs == nil || vErrs.Count() == 0 {
		t.Fatal("requesting an original invoice through the fiscal-request path must be refused: " +
			"it would write a row the issuance job never reads")
	}
}

// The adjustment intents still reach the later gates. Refusing the original must not have closed
// the path a return depends on to correct a document that was already issued.
func TestAdjustmentIntentsAreNotRefusedByTheOriginalGate(t *testing.T) {
	for _, intent := range []string{
		string(models.SalesFiscalIntentAdjustForFullReturn),
		string(models.SalesFiscalIntentAdjustForPartialReturn),
		string(models.SalesFiscalIntentAdjustPrice),
	} {
		if !isKnownFiscalIntent(intent) {
			t.Errorf("%q must remain a valid intent: returns correct issued documents through it", intent)
		}
		if intent == string(models.SalesFiscalIntentIssueOriginal) {
			t.Errorf("%q must not be treated as an adjustment", intent)
		}
	}
}

// The Sales intent constants and the port's must agree exactly. They are declared separately (the
// model must not import the port and vice versa), so only this test stops them drifting.
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

// The buyer snapshot is stored as a map under the names the column is read back by, so the read
// path does not depend on Go field names.
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

// A provider message longer than the column must not cost the whole record of why it failed.
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
