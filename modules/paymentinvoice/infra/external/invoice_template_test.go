package external

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The prefix must pass Essential's own template rules and render the number shape invoices have
// always carried. Checked here, the one package allowed to know Essential's internals.
func TestTheInvoicePrefixRendersTodaysShape(t *testing.T) {
	cmd := services.InvoiceSequenceCommand("01ORG")
	prefix := *cmd.PrefixTemplate

	require.NoError(t, seqalgo.ValidateTemplate(prefix, *cmd.FormatOwner, *cmd.ResetPeriod, true))
	require.NoError(t, seqalgo.ValidatePeriodMarker(prefix, "", *cmd.ResetPeriod))

	at := time.Date(2026, 11, 4, 9, 0, 0, 0, time.UTC)
	period, err := seqalgo.PeriodOf(at, "UTC", *cmd.ResetPeriod)
	require.NoError(t, err)
	rendered, err := seqalgo.RenderTemplate(prefix, seqalgo.TemplateVars{
		EffectiveAt: at, AcceptedAt: at, RangeStart: period.Start, PeriodKey: period.Key, Location: time.UTC,
	})
	require.NoError(t, err)
	assert.Equal(t, "2026", period.Key)
	assert.Equal(t, "INV-2026-000001", fmt.Sprintf("%s%0*d", rendered, *cmd.Padding, 1))
}
