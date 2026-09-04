// Package cqrs receives the commands the scheduler dispatches into Sales.
//
// A handler here is thin by design: it resolves the settings, calls the domain, and turns the result
// into a reply. The rules about what may be issued live in domain/services, so that a run triggered
// by the scheduler and one triggered by hand behave identically.
package cqrs

import (
	"context"
	"sync"
	"time"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// issuancePageSize bounds one pass.
//
// It exists because the scheduler BLOCKS ON THE REPLY: the dispatch is a request/reply over the
// command bus with a timeout in the tens of seconds, so a pass that tried to drain a large backlog
// would time out, be recorded as failed, and be retried — doing the same slow work again. A bounded
// pass finishes, reports, and lets the next tick take the rest.
const issuancePageSize = 50

// jobNameEinvoiceIssuance labels this pass in the tenant fanout's logs, which name the job whose
// sweep failed for a tenant. Only a log label: the scheduler identifies the job by its job_key.
const jobNameEinvoiceIssuance = "sales einvoice issuance"

type BillingHandler struct {
	invoicing itInvoicing.InvoicingExtService
	parties   itExt.PartyExtService
	settings  itExt.EffectiveSettingsExtService
	logger    logging.LoggerService
}

func NewBillingHandler(
	invoicing itInvoicing.InvoicingExtService,
	parties itExt.PartyExtService,
	settings itExt.EffectiveSettingsExtService,
	logger logging.LoggerService,
) *BillingHandler {
	return &BillingHandler{
		invoicing: invoicing,
		parties:   parties,
		settings:  settings,
		logger:    logger,
	}
}

// IssueEinvoices runs one issuance pass.
//
// It returns an error only when the pass could not run at all. A single instruction the provider
// refused is data, not a failure: reporting it as one would have the scheduler retry the whole pass,
// re-attempting every other instruction to get at the one that will be refused again.
func (this *BillingHandler) IssueEinvoices(
	ctx context.Context, packet *cqrs.RequestPacket[itBilling.IssueEinvoicesCommand],
) (*cqrs.Reply[itBilling.IssueEinvoicesResult], error) {
	limit := issuancePageSize
	if request := packet.Request(); request != nil && request.Limit > 0 {
		limit = request.Limit
	}

	// The scheduler's jobs are global: nothing about a cron tick names a tenant, and in coremart every
	// tenant-scoped read from a context that names none PANICS rather than erroring. So the pass is
	// written as a sweep that never mentions tenancy, and ScopeSweep supplies the scoping the
	// deployment installed — once per enabled tenant in coremart, a single plain run in nikkierp.
	//
	// Deliberately not `ctx.(corectx.Context)`: the bus does build a request context, but it is the
	// tenantless one the dispatch arrived on, and using it is what made this pass fail on every tick.
	var mutex sync.Mutex
	total := &services.IssueEinvoicesResult{}
	now := time.Now().UTC()

	sweep := func(tenantCtx corectx.Context) error {
		policy := services.ResolveSalesPolicy(tenantCtx, this.settings)

		result, err := services.IssueDueEinvoices(
			tenantCtx, this.invoicing, this.parties, policy, now, limit)
		if err != nil {
			return err
		}

		// The fanout may run sweeps concurrently, and each tenant's counts land in one shared total.
		mutex.Lock()
		defer mutex.Unlock()
		total.Examined += result.Examined
		total.Issued += result.Issued
		total.Failed += result.Failed
		total.Indeterminate += result.Indeterminate
		return nil
	}

	if err := job.ScopeSweep(sweep, jobNameEinvoiceIssuance)(ctx, nil); err != nil {
		return nil, err
	}
	result := total

	if this.logger != nil && result.Examined > 0 {
		this.logger.Infof(
			"sales einvoice issuance: examined %d, issued %d, failed %d, indeterminate %d",
			result.Examined, result.Issued, result.Failed, result.Indeterminate)
	}

	// Raised rather than counted quietly: an indeterminate attempt means a document may exist that
	// nothing here knows about, and no amount of waiting resolves it.
	if this.logger != nil && result.Indeterminate > 0 {
		this.logger.Warnf(
			"sales einvoice issuance: %d attempts left undetermined and need reconciling",
			result.Indeterminate)
	}

	return &cqrs.Reply[itBilling.IssueEinvoicesResult]{
		Result: itBilling.IssueEinvoicesResult{
			Examined:      result.Examined,
			Issued:        result.Issued,
			Failed:        result.Failed,
			Indeterminate: result.Indeterminate,
		},
	}, nil
}

// InitCqrsHandlers subscribes Sales' command handlers.
//
// It must run BEFORE the job is registered in OnAppStarted: the scheduler validates that a job's
// command name is a registered request type, and rejects the registration otherwise.
func InitCqrsHandlers() error {
	if err := deps.Register(NewBillingHandler); err != nil {
		return err
	}
	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *BillingHandler) error {
		return cqrsBus.SubscribeRequests(
			context.Background(),
			cqrs.NewHandler(handler.IssueEinvoices),
		)
	})
}
