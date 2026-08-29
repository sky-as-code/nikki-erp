package app

import (
	"context"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// The outbox sweep. Domain services write the outbox row inside their own transaction; this sweep
// moves rows to the broker afterwards, turning an atomicity problem into a retryable delivery one.
// It publishes and THEN marks: marking first would lose an event whose publish failed, whereas
// marking after can only duplicate one, which consumers must already deduplicate on event_id. The
// cron registry applies no distributed lock, so two instances duplicate work but corrupt nothing.

const (
	// Every minute: stock, accounting and the kiosk fleet wait on these events, and an empty sweep
	// costs one indexed query.
	cronOutboxSweep = "* * * * *"

	jobNameOutboxSweep = "sales-outbox-sweep"

	// outboxPageSize bounds one sweep: small enough that a backlog drains over several runs rather
	// than one long transaction, large enough for a normal minute in a single pass.
	outboxPageSize = 200
)

// OutboxJobs drains the integration event outbox.
type OutboxJobs struct {
	publisher itMessage.IntegrationEventPublisher
	logger    logging.LoggerService
}

func NewOutboxJobs(
	publisher itMessage.IntegrationEventPublisher, logger logging.LoggerService,
) *OutboxJobs {
	return &OutboxJobs{publisher: publisher, logger: logger}
}

func (this *OutboxJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronOutboxSweep, jobNameOutboxSweep,
		job.ScopeSweep(this.Sweep, jobNameOutboxSweep))
}

// Sweep publishes the events waiting for the broker. One event's failure does not stop the sweep,
// otherwise a single bad row would strand every event behind it on every run.
func (this *OutboxJobs) Sweep(ctx corectx.Context) error {
	if this.publisher == nil {
		// No broker adapter in this build; rows accumulate unpublished, which is the honest state.
		return nil
	}

	pending, err := services.UnpublishedEvents(ctx, outboxPageSize)
	if err != nil {
		return err
	}

	// A plain context for the publish: the producing request committed long ago, and a retry may run
	// hours later.
	publishCtx := context.Background()

	for _, row := range pending {
		event := services.IntegrationEventOf(row)

		if err := this.publisher.Publish(publishCtx, event.IntegrationEvent); err != nil {
			this.logError("sales outbox: publishing event failed", event.EventId, err)

			// The row stays unpublished so the next sweep retries it. A failure to record the failure
			// only loses the attempt count, so it is logged rather than abandoning the page.
			if recordErr := services.RecordPublishFailure(ctx, row, err.Error()); recordErr != nil {
				this.logError("sales outbox: recording a publish failure failed",
					event.EventId, recordErr)
			}
			continue
		}

		if err := services.MarkEventPublished(ctx, event.RowId); err != nil {
			// Published but not marked: the next sweep republishes it, and consumer event_id
			// deduplication makes that harmless.
			this.logError("sales outbox: event published but not marked", event.EventId, err)
		}
	}
	return nil
}

// logError puts the event id in the message text because the logger takes an error, not a field
// map, and the id is what an operator greps the outbox for.
func (this *OutboxJobs) logError(message, eventId string, err error) {
	if this.logger == nil {
		return
	}
	this.logger.Error(message+" (event "+eventId+")", err)
}
