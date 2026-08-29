package app

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The expiry sweep. A stale draft holds voucher (and later stock) reservations that block other
// customers, so it is expired to release them. It runs hourly because the window is measured in
// hours (draft_order_expiry_hours, default 24). The cron registry applies no distributed lock, but
// each expiry re-reads the order in its own transaction and moves it only from `draft`, so a second
// instance cannot expire it twice.

const (
	// Five past the hour, offset so it does not contend with every other hourly job.
	cronExpirySweep = "5 * * * *"

	jobNameExpirySweep = "sales-expiry-sweep"

	// expiryPageSize bounds one sweep so a backlog drains over several runs instead of one long pass
	// holding a connection. Each expiry is its own transaction, so a bounded run stays committed.
	expiryPageSize = 200
)

// ExpiryJobs expires stale drafts and lapsed quotations.
type ExpiryJobs struct {
	settings itExt.EffectiveSettingsExtService
	logger   logging.LoggerService

	// now is injected so a test can drive the clock.
	now func() time.Time
}

func NewExpiryJobs(
	settings itExt.EffectiveSettingsExtService, logger logging.LoggerService,
) *ExpiryJobs {
	return &ExpiryJobs{settings: settings, logger: logger, now: time.Now}
}

func (this *ExpiryJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronExpirySweep, jobNameExpirySweep,
		job.ScopeSweep(this.Sweep, jobNameExpirySweep))
}

// Sweep expires what has gone stale. Orders and quotations are swept independently: a failure on
// one must not stop the other.
func (this *ExpiryJobs) Sweep(ctx corectx.Context) error {
	now := this.now().UTC()
	policy := services.ResolveSalesPolicy(ctx, this.settings)

	orders, err := services.ExpireStaleDrafts(ctx, policy, now, expiryPageSize)
	if err != nil {
		this.logError("sales expiry: expiring stale drafts failed", err)
	} else if len(orders.ExpiredOrderIds) > 0 {
		this.logInfo("sales expiry: expired drafts", len(orders.ExpiredOrderIds),
			len(orders.ReleasedVoucherCodeIds))
	}

	quotations, err := services.ExpireLapsedQuotations(ctx, now, expiryPageSize)
	if err != nil {
		this.logError("sales expiry: expiring lapsed quotations failed", err)
		return nil
	}
	if len(quotations.ExpiredQuotationIds) > 0 {
		this.logInfo("sales expiry: expired quotations",
			len(quotations.ExpiredQuotationIds), 0)
	}
	return nil
}

func (this *ExpiryJobs) logError(message string, err error) {
	if this.logger == nil {
		return
	}
	this.logger.Error(message, err)
}

// logInfo reports what the sweep did. The released-code count is logged separately because it is
// the part with an effect outside these orders, and what an investigator greps for.
func (this *ExpiryJobs) logInfo(message string, count, releasedCodes int) {
	if this.logger == nil {
		return
	}
	this.logger.Infof("%s: %d, released voucher codes: %d", message, count, releasedCodes)
}
