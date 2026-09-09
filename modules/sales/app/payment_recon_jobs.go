package app

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The payment reconciliation sweep, which is what makes a gateway payment safe to rely on.
//
// A settlement is announced on the event bus, and that bus acknowledges a message before a
// subscriber has handled it. A crash between the two loses the announcement, and a customer who has
// paid would be left with an open bill and no way to close it. This sweep asks the gateway directly
// for any payment that has waited too long, so the worst case becomes a few minutes' delay rather
// than money nobody can account for.
//
// Every five minutes, not every minute: it costs a round trip per stale payment, and a customer
// whose settlement was lost is already waiting on a human timescale.

const (
	cronPaymentReconSweep = "*/5 * * * *"

	jobNamePaymentReconSweep = "sales-payment-recon-sweep"

	// paymentReconPageSize bounds one pass so a backlog drains over several runs rather than one
	// long pass holding a connection. Each verdict is applied in its own write, so a bounded run
	// stays committed.
	paymentReconPageSize = 100

	// paymentReconMinAge keeps the sweep off payments that were only just opened: the customer may
	// still be at the terminal, and asking about every fresh order would be a round trip for an
	// answer nobody has decided yet. It is comfortably longer than a normal QR payment takes.
	paymentReconMinAge = 10 * time.Minute

	// fiscalReconMinAge is longer: issuing a document is slower than settling a payment, and a
	// request still in flight must not be asked about as though it were lost.
	fiscalReconMinAge = 30 * time.Minute
)

// PaymentReconJobs settles the payments whose verdict was never announced, and resolves the fiscal
// requests whose answer never came back.
//
// The two run together because they are the same failure: Sales asked another party for something,
// and the reply was lost. Neither can be resolved by waiting, and both are safe to ask about
// repeatedly — which is what lets one sweep cover them.
type PaymentReconJobs struct {
	orders    itExt.PaymentOrderExtService
	invoicing itInvoicing.InvoicingExtService
	logger    logging.LoggerService

	// now is injected so a test can drive the clock.
	now func() time.Time
}

func NewPaymentReconJobs(
	orders itExt.PaymentOrderExtService,
	invoicing itInvoicing.InvoicingExtService,
	logger logging.LoggerService,
) *PaymentReconJobs {
	return &PaymentReconJobs{
		orders: orders, invoicing: invoicing, logger: logger, now: time.Now,
	}
}

func (this *PaymentReconJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronPaymentReconSweep, jobNamePaymentReconSweep,
		job.ScopeSweep(this.Sweep, jobNamePaymentReconSweep))
}

// Sweep applies the verdicts that were never announced.
//
// It returns nil even when the pass failed: the cron registry has no retry, and the next tick is
// five minutes away regardless. A failure is logged, and the same payments are picked up then.
func (this *PaymentReconJobs) Sweep(ctx corectx.Context) error {
	// Always runs, whatever the payments did: the two are independent failures of the same kind.
	defer this.sweepFiscal(ctx)

	result, err := services.ReconcileStalePayments(
		ctx, this.orders, this.now().UTC(), paymentReconMinAge, paymentReconPageSize)
	if err != nil {
		if this.logger != nil {
			this.logger.Error("sales payment reconciliation: sweep failed", err)
		}
		return nil
	}

	if this.logger == nil || result.Examined == 0 {
		return nil
	}

	// Logged only when it did something. A quiet sweep is the normal case — the announcements
	// usually arrive — and logging every empty pass would bury the ones that mattered.
	if result.Settled > 0 || result.Failed > 0 || result.Unknown > 0 {
		this.logger.Infof(
			"sales payment reconciliation: examined %d, settled %d, failed %d, unknown %d",
			result.Examined, result.Settled, result.Failed, result.Unknown)
	}

	// An order Sales holds a correlation for but paymentinvoice has never heard of cannot be fixed
	// by waiting, so it is raised rather than counted quietly.
	if result.Unknown > 0 {
		this.logger.Warnf(
			"sales payment reconciliation: %d pending payments name a payment order that does not exist",
			result.Unknown)
	}
	return nil
}

// sweepFiscal resolves the fiscal requests whose answer was lost.
//
// Run after the payments and independently: a failure reading one must not stop the other, and a
// document is not more or less issued because a payment could not be reconciled.
func (this *PaymentReconJobs) sweepFiscal(ctx corectx.Context) {
	result, err := services.ReconcileStaleFiscalRequests(
		ctx, this.invoicing, this.now().UTC(), fiscalReconMinAge, paymentReconPageSize)
	if err != nil {
		if this.logger != nil {
			this.logger.Error("sales fiscal reconciliation: sweep failed", err)
		}
		return
	}

	if this.logger == nil || result.Resolved == 0 {
		return
	}
	this.logger.Infof(
		"sales fiscal reconciliation: examined %d, resolved %d, still unanswered %d",
		result.Examined, result.Resolved, result.Unresolved)
}
