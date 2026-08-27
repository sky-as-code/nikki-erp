package app

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The expiry sweep (BR 87.3, SALES-040).
//
// # Why this runs at all
//
// A stale draft is not merely untidy. It holds voucher reservations that stop another customer using
// a code, and — once inventory reservation is wired — stock that stops the goods being sold. Both are
// held on the promise that the sale is about to happen, and a draft nobody has touched for a day has
// stopped promising it. BR 87.3 justifies the feature by future eCommerce; the reservations are why
// it is worth having now.
//
// # Hourly, not every minute
//
// The window is measured in HOURS (draft_order_expiry_hours, default 24), so a sweep every minute
// would do the same work sixty times to change nothing. An order expiring up to an hour late costs
// nobody anything; the reservation it holds was already a day old.
//
// # Two instances are safe
//
// The cron registry applies no distributed lock. Each expiry re-reads the order inside its own
// transaction and moves it only from `draft`, so a second instance meeting an already-expired order
// finds nothing to do rather than expiring it twice.

const (
	// cronExpirySweep runs at five past every hour. Offset from the hour so it does not contend
	// with every other hourly job in the deployment at exactly the same moment.
	cronExpirySweep = "5 * * * *"

	jobNameExpirySweep = "sales-expiry-sweep"

	// expiryPageSize bounds one sweep. A backlog is drained over several runs rather than in one
	// long pass holding a connection — and since each expiry is its own transaction, a bounded run
	// leaves the work it did committed rather than rolling it all back.
	expiryPageSize = 200
)

// ExpiryJobs expires stale drafts and lapsed quotations.
type ExpiryJobs struct {
	settings itExt.EffectiveSettingsExtService
	logger   logging.LoggerService

	// now is injected so a test can drive the clock. Production passes time.Now.
	now func() time.Time
}

func NewExpiryJobs(
	settings itExt.EffectiveSettingsExtService, logger logging.LoggerService,
) *ExpiryJobs {
	return &ExpiryJobs{settings: settings, logger: logger, now: time.Now}
}

// RegisterJobs puts the sweep on the scheduler.
func (this *ExpiryJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronExpirySweep, jobNameExpirySweep,
		job.ScopeSweep(this.Sweep, jobNameExpirySweep))
}

// Sweep expires what has gone stale.
//
// Orders and quotations are swept independently: a failure reading one must not stop the other, since
// they share nothing but a schedule and a lapsed quotation is no less lapsed because an order could
// not be read.
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

// logInfo reports what the sweep did, including the released codes.
//
// The code count is worth a line of its own: it is the part with an effect outside these orders —
// another customer can now use those codes — and it is what somebody investigating "why did this
// voucher become available again" will search for.
func (this *ExpiryJobs) logInfo(message string, count, releasedCodes int) {
	if this.logger == nil {
		return
	}
	this.logger.Infof("%s: %d, released voucher codes: %d", message, count, releasedCodes)
}
