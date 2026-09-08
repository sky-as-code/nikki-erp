package app

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The reservation expiry sweep, Inventory's first cron.
//
// A hold taken for a customer-controlled demand — an order to be collected at a kiosk, say — keeps
// stock out of the sellable pool for as long as it stands. Without a sweep, one order the customer
// never came back for would make a slot permanently unsellable, and the goods would sit there
// visible to nobody.
//
// What this job does is narrow on purpose: it releases stock and cancels the hold document, and
// nothing else. Expiring a hold does not cancel the demand behind it and refunds nobody — those are
// the owning module's decisions, and taking them here would refund a customer who is still walking
// to the machine. The owning module notices its reservation has lapsed and decides what to do.
//
// The cron registry applies no distributed lock. It does not need one: each release re-reads its
// transfer and acts only on one still open, so a second instance sweeping the same row finds
// nothing left to do rather than releasing twice.

const (
	// Twenty past the hour, offset away from the other hourly sweeps so they do not contend.
	cronReservationExpiry = "20 * * * *"

	jobNameReservationExpiry = "inventory-reservation-expiry-sweep"

	// expiryPageSize bounds one pass so a backlog drains over several runs rather than one long
	// pass holding a connection. Each release commits on its own, so a bounded run stays done.
	expiryPageSize = 200
)

// ReservationExpiryJobs releases holds whose time has run out.
type ReservationExpiryJobs struct {
	transfers itStock.StockTransferMovementService
	logger    logging.LoggerService

	// now is injected so a test can drive the clock.
	now func() time.Time
}

func NewReservationExpiryJobs(
	transfers itStock.StockTransferMovementService, logger logging.LoggerService,
) *ReservationExpiryJobs {
	return &ReservationExpiryJobs{transfers: transfers, logger: logger, now: time.Now}
}

func (this *ReservationExpiryJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronReservationExpiry, jobNameReservationExpiry,
		job.ScopeSweep(this.Sweep, jobNameReservationExpiry))
}

// Sweep releases every hold that has lapsed as of now.
func (this *ReservationExpiryJobs) Sweep(ctx corectx.Context) error {
	released, err := this.transfers.ExpireLapsedReservations(ctx, this.now().UTC(), expiryPageSize)
	if err != nil {
		return err
	}
	if released > 0 && this.logger != nil {
		this.logger.Infof("released %d lapsed stock reservation(s)", released)
	}
	return nil
}
