package sales

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The electronic-invoice job's registration.
//
// It goes to the scheduler module rather than the in-process cron that runs Sales' other sweeps, and
// the difference is what the work produces. A sweep that expires stale drafts only touches Sales'
// own rows, so a missed tick costs nothing — the next one finds the same rows. Issuing an electronic
// invoice calls a third party and produces a legal document, so a run that failed has to be visibly
// failed, retried on a policy, and auditable afterwards.

const (
	// Every ten minutes, in UTC. A cron pinned to local time would drift by an hour twice a year,
	// and a job that issues invoices running twice in one hour is worse than one running at an odd
	// local time.
	cronEinvoiceIssuance = "*/10 * * * *"

	jobKeyEinvoiceIssuance  = "einvoice-issuance"
	jobNameEinvoiceIssuance = "Sales electronic invoice issuance"

	// The command the scheduler dispatches. It must match the request type the CQRS handler
	// subscribed with during Init, or registration is refused.
	commandEinvoiceIssuance = "sales_billing.issueEinvoices"

	// Three tries, a minute apart. The failures worth retrying here are transient — the provider
	// unreachable, a database blip — and anything that survives three attempts needs a human rather
	// than a fourth. Each attempt re-reads what is due, so a retry never repeats work that succeeded.
	einvoiceMaxAttempts          = int32(3)
	einvoiceRetryIntervalSeconds = int32(60)
)

// registerEinvoiceJob ensures the recurring issuance job exists.
//
// Idempotent: it runs on every boot and the second one is a no-op. It does NOT update a job that is
// already registered, so changing the schedule is a deliberate act against the scheduler rather than
// a side effect of a redeploy — which is what stops a rolled-back release quietly reverting an
// operator's change.
func registerEinvoiceJob(
	ctx corectx.Context, scheduler itExt.SchedulerExtService, logger logging.LoggerService,
) error {
	if scheduler == nil {
		return nil
	}

	result, err := scheduler.EnsureJob(ctx, itExt.EnsureJobCommand{
		ModuleName:           "sales",
		JobKey:               jobKeyEinvoiceIssuance,
		Name:                 jobNameEinvoiceIssuance,
		CronExpression:       cronEinvoiceIssuance,
		CommandName:          commandEinvoiceIssuance,
		MaxAttempts:          einvoiceMaxAttempts,
		RetryIntervalSeconds: einvoiceRetryIntervalSeconds,
	})
	if err != nil {
		return err
	}

	// Logged only on the first registration. Every later boot re-registers the same job, and saying
	// so each time would bury the one occurrence that is actually news.
	if logger != nil && result != nil && result.WasCreated {
		logger.Infof("sales: registered the electronic invoice issuance job (%s)", result.JobId)
	}
	return nil
}
