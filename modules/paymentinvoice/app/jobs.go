package app

import (
	stdErr "errors"
	"time"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The schedules.
//
// The watchdog runs often because the window between a lost callback and a customer complaining is
// short. The cleaner runs once a day because it only deletes what nobody is going to ask about.
//
// Single instance is assumed, inherited from the service this replaces (its process manager pinned
// one instance). Two instances would run every sweep twice; that is not corrupting — each write
// re-checks the order's state inside its own transaction, so the second run finds nothing to do —
// but it is wasted work, and distributed locking is out of scope until the deployment needs it.
const (
	cronWatchdog  = "* * * * *"
	cronCleaner   = "0 0 * * *"
	cronSyncRetry = "*/5 * * * *"

	jobNameWatchdog  = "paymentinvoice-watchdog"
	jobNameCleaner   = "paymentinvoice-cleaner"
	jobNameSyncRetry = "paymentinvoice-sync-retry"
)

// syncGracePeriod is how long a just-settled order is left alone before the retry sweep considers
// it un-notified.
//
// It is comfortably longer than one full send — the client's per-attempt timeout times its
// retries, plus the backoff between them — so a notification still in flight is not duplicated by
// a sweep that happens to tick while it runs. It is a constant rather than configuration because
// it is derived from the client's own bounds, not from anything a deployment chooses.
const syncGracePeriod = 2 * time.Minute

// JobsConfig is the tuning the sweeps read from configuration.
type JobsConfig struct {
	// ExpireAfter is how long an order may sit unpaid before the gateway is asked for a verdict.
	ExpireAfter time.Duration

	// CleanAfter is how long an unpaid or expired order is kept before deletion.
	CleanAfter time.Duration
}

// JobsManager runs the module's background sweeps.
type JobsManager struct {
	orders   *services.OrderDomainService
	notifier *ResultNotifier
	config   JobsConfig
	logger   logging.LoggerService

	// now is injected so the sweeps can be tested against a fixed clock rather than by waiting.
	now func() time.Time
}

func NewJobsManager(
	orders *services.OrderDomainService,
	notifier *ResultNotifier,
	config JobsConfig,
	logger logging.LoggerService,
) *JobsManager {
	return &JobsManager{
		orders:   orders,
		notifier: notifier,
		config:   config,
		logger:   logger,
		now:      time.Now,
	}
}

// RegisterJobs puts the three sweeps on the scheduler.
func (this *JobsManager) RegisterJobs(registry job.CronjobRegistry) error {
	// Through ScopeSweep rather than the local wrap(): in a multi-tenant build each sweep must run
	// once per tenant, and a sweep handed a tenantless context panics on its first read. See the
	// package comment on core/job/sweep_scope.go.
	return stdErr.Join(
		registry.Register(cronWatchdog, jobNameWatchdog,
			job.ScopeSweep(this.Watchdog, jobNameWatchdog)),
		registry.Register(cronCleaner, jobNameCleaner,
			job.ScopeSweep(this.Cleaner, jobNameCleaner)),
		registry.Register(cronSyncRetry, jobNameSyncRetry,
			job.ScopeSweep(this.SyncRetry, jobNameSyncRetry)),
	)
}

// Watchdog asks the gateway about orders no callback ever arrived for.
//
// This is the recovery path for a lost callback, and it matters because a gateway that could not
// deliver its callback still took the customer's money. An order the gateway confirms as paid is
// settled here exactly as the callback would have settled it; one it has reached no verdict on is
// left alone, because a customer still standing at a terminal has not failed to pay; one that is
// past its window with no verdict is expired.
//
// One order's failure does not stop the sweep. An unreachable gateway or a single bad row would
// otherwise strand every order behind it in the page, and the next run would meet the same row
// first and strand them again.
func (this *JobsManager) Watchdog(ctx corectx.Context) error {
	cutoff := this.now().Add(-this.config.ExpireAfter)

	stale, err := services.FindStaleOrders(ctx, cutoff)
	if err != nil {
		return errors.Wrap(err, jobNameWatchdog)
	}
	if len(stale) == 0 {
		return nil
	}
	this.logger.Infof("%s: %d order(s) awaiting a verdict", jobNameWatchdog, len(stale))

	for _, order := range stale {
		this.reconcileOne(ctx, order)
	}
	return nil
}

// reconcileOne settles or expires a single stale order, logging rather than propagating a failure.
func (this *JobsManager) reconcileOne(ctx corectx.Context, order services.StaleOrder) {
	verdict, err := this.orders.ReconcileStaleOrder(ctx, order)
	if err != nil {
		// The gateway could not be asked. The order keeps its status and the next run asks again:
		// an unreachable gateway is not evidence about a payment.
		this.logger.Warnf("%s: order '%s' could not be checked: %s",
			jobNameWatchdog, order.OrderId, err.Error())
		return
	}

	if verdict.Settled {
		if verdict.Applied {
			this.notify(ctx, order, verdictStatus(verdict.Paid))
		}
		// Not applied means a callback settled it first, and that path sent its own notification.
		return
	}

	expired, err := this.orders.ExpireStaleOrder(ctx, order)
	if err != nil {
		this.logger.Errorf("%s: order '%s' could not be expired: %s",
			jobNameWatchdog, order.OrderId, err.Error())
		return
	}
	if expired {
		this.notify(ctx, order, services.OrderStatusExpiredForSync)
	}
}

// Cleaner deletes orders old enough that nobody will ask about them.
//
// Only unpaid and expired orders are in scope. A paid, failed, canceled or refunded order is the
// financial record of a transaction and is never deleted by a sweep.
func (this *JobsManager) Cleaner(ctx corectx.Context) error {
	cutoff := this.now().Add(-this.config.CleanAfter)

	cleanable, err := services.FindCleanableOrders(ctx, cutoff)
	if err != nil {
		return errors.Wrap(err, jobNameCleaner)
	}
	if len(cleanable) == 0 {
		return nil
	}

	deleted := 0
	for _, order := range cleanable {
		if err := services.DeleteOrder(ctx, order.Pk); err != nil {
			this.logger.Warnf("%s: order '%s' could not be deleted: %s",
				jobNameCleaner, order.OrderId, err.Error())
			continue
		}
		deleted++
	}
	this.logger.Infof("%s: deleted %d of %d stale order(s)", jobNameCleaner, deleted, len(cleanable))
	return nil
}

// SyncRetry re-notifies the ordering system about payments it was never told of.
//
// The service this module replaces declared this job and never implemented it, so a tenant that
// was down when a payment settled never learned of it at all. The retry is bounded by the client's
// own attempt count; an order whose notification keeps failing stops being retried and stays
// visibly failed rather than being retried forever.
//
// Orders settled within the grace period are left for the next run: a callback sends its
// notification off the request, so one that has only just settled probably has a send in flight.
func (this *JobsManager) SyncRetry(ctx corectx.Context) error {
	pending, err := services.FindOrdersNeedingSync(ctx, this.now().Add(-syncGracePeriod))
	if err != nil {
		return errors.Wrap(err, jobNameSyncRetry)
	}
	if len(pending) == 0 {
		return nil
	}
	this.logger.Infof("%s: %d order(s) to re-notify", jobNameSyncRetry, len(pending))

	for _, order := range pending {
		this.notify(ctx, services.StaleOrder{
			Pk:        order.Pk,
			OrderId:   order.OrderId,
			ReturnUrl: order.ReturnUrl,
		}, order.Status)
	}
	return nil
}

// notify reports one order, blocking until the send is done.
//
// A sweep notifies inline rather than detaching, unlike the gateway callbacks: nothing is waiting
// on it, and working through a page one order at a time is what stops a backlog from opening a
// goroutine and a connection for every order at once.
func (this *JobsManager) notify(ctx corectx.Context, order services.StaleOrder, status string) {
	this.notifier.Notify(ctx, NotifyTarget{
		Pk:        order.Pk,
		OrderId:   order.OrderId,
		ReturnUrl: order.ReturnUrl,
	}, status)
}

// verdictStatus turns a paid/not-paid verdict into the status the ordering system is told.
func verdictStatus(paid bool) string {
	if paid {
		return services.OrderStatusPaidForSync
	}
	return services.OrderStatusFailedForSync
}
