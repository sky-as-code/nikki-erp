package app

import (
	"context"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// The outbox sweep (BR 80, acceptance 94.34, SALES-037).
//
// # Why a sweep rather than publishing inline
//
// A domain service writes the outbox row inside its own transaction, so the row and the business
// change cannot disagree. This sweep is the other half: it moves rows to the broker afterwards,
// which is what turns an atomicity problem — publish-then-commit announces a sale that rolls back,
// commit-then-publish loses one that happened — into a delivery problem that retrying fixes.
//
// # Publish, THEN mark
//
// The ordering is the deliberate half of the at-least-once trade. Marking first would lose any event
// whose publish then failed; marking after can only republish one, and a duplicate is precisely what
// the consumer's event_id deduplication already handles.
//
// # Two instances are safe, but wasteful
//
// The cron registry applies no distributed lock, so two instances run every sweep twice and may
// publish the same event twice. That is not corrupting — consumers deduplicate on event_id, which is
// required of them anyway by acceptance 94.34 — but it is wasted work. Locking is out of scope until
// the deployment needs it, and the same assumption paymentinvoice's sweeps make.

const (
	// cronOutboxSweep runs every minute. The window between a sale confirming and a consumer needing
	// to know about it is short — stock, accounting and the kiosk fleet all wait on it — and a
	// sweep that finds nothing costs one indexed query against a partial-index-shaped predicate.
	cronOutboxSweep = "* * * * *"

	jobNameOutboxSweep = "sales-outbox-sweep"

	// outboxPageSize bounds one sweep. Small enough that a backlog is drained over several runs
	// rather than in one long transaction holding a connection, and large enough that a normal
	// minute's events go in a single pass.
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

// RegisterJobs puts the sweep on the scheduler.
func (this *OutboxJobs) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronOutboxSweep, jobNameOutboxSweep,
		job.ScopeSweep(this.Sweep, jobNameOutboxSweep))
}

// Sweep publishes the events waiting for the broker.
//
// One event's failure does not stop the sweep. An unreachable broker or a single unserialisable
// payload would otherwise strand every event behind it in the page, and the next run would meet the
// same row first and strand them again — so a failure is counted against its own row and the sweep
// moves on.
func (this *OutboxJobs) Sweep(ctx corectx.Context) error {
	if this.publisher == nil {
		// No broker adapter bound in this build. Rows accumulate unpublished, which is the honest
		// state: the events really happened and nothing has carried them anywhere.
		return nil
	}

	pending, err := services.UnpublishedEvents(ctx, outboxPageSize)
	if err != nil {
		return err
	}

	// A plain context for the publish itself, matching the port: the request that produced these
	// events committed and returned long ago, and a retry may be running hours later.
	publishCtx := context.Background()

	for _, row := range pending {
		event := services.IntegrationEventOf(row)

		if err := this.publisher.Publish(publishCtx, event.IntegrationEvent); err != nil {
			this.logError("sales outbox: publishing event failed", event.EventId, err)

			// The row stays unpublished so the next sweep retries it. A failure to RECORD the
			// failure is logged and otherwise ignored: it would only lose the attempt count, and
			// abandoning the rest of the page over it would strand events that could have gone.
			if recordErr := services.RecordPublishFailure(ctx, row, err.Error()); recordErr != nil {
				this.logError("sales outbox: recording a publish failure failed",
					event.EventId, recordErr)
			}
			continue
		}

		if err := services.MarkEventPublished(ctx, event.RowId); err != nil {
			// Published but not marked. The next sweep will publish it again, and the consumer's
			// event_id deduplication is what makes that harmless — this is the case at-least-once
			// delivery exists to absorb, so it is logged rather than escalated.
			this.logError("sales outbox: event published but not marked", event.EventId, err)
		}
	}
	return nil
}

// logError reports a failure against the event it belongs to.
//
// The event id goes into the message rather than into structured fields, because the logger takes an
// error and not a field map - and an id is the one thing that makes a broker failure actionable,
// since it is what an operator greps the outbox for.
func (this *OutboxJobs) logError(message, eventId string, err error) {
	if this.logger == nil {
		return
	}
	this.logger.Error(message+" (event "+eventId+")", err)
}
