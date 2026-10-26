package sales

import (
	"testing"

	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
)

// The job's command name and the handler's request type are declared in two places and must agree.
//
// If they drift, nothing fails loudly: the scheduler refuses the registration at boot with an error
// nobody is watching, and the symptom is invoices silently never being issued. This is the cheapest
// place to catch that.
func TestTheEinvoiceJobDispatchesToARegisteredCommand(t *testing.T) {
	want := itBilling.IssueEinvoicesCommand{}.CqrsRequestType().String()
	if commandEinvoiceIssuance != want {
		t.Errorf("the job dispatches %q but the handler subscribes %q; the scheduler would "+
			"refuse this registration and no invoice would ever be issued",
			commandEinvoiceIssuance, want)
	}
}

// A ten-minute cron in the five-field form the scheduler parses. Pinned because the field count is
// the easy mistake — a six-field spec with seconds is what most people write from memory, and it
// would be rejected at registration.
func TestTheEinvoiceCronIsFiveFields(t *testing.T) {
	fields := 1
	for _, char := range cronEinvoiceIssuance {
		if char == ' ' {
			fields++
		}
	}
	if fields != 5 {
		t.Errorf("cron %q has %d fields, want 5", cronEinvoiceIssuance, fields)
	}
}

// The retry policy must count more than one try, or a transient provider outage would leave the
// pass failed until the next tick with no attempt to recover in between.
func TestTheEinvoiceJobRetries(t *testing.T) {
	if einvoiceMaxAttempts < 2 {
		t.Errorf("max attempts = %d, which means no retry at all", einvoiceMaxAttempts)
	}
	if einvoiceRetryIntervalSeconds <= 0 {
		t.Error("a retry interval of zero would hammer a provider that is already struggling")
	}
}
