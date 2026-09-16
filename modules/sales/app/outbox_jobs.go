package app

import (
	"context"
	"fmt"

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

	// drainPageSize bounds one immediate drain. Much smaller than a sweep's page on purpose: a nudge
	// fires once per write, so several may run at once, and each is meant to carry the event that
	// just happened rather than take on the backlog the sweep exists to clear.
	drainPageSize = 20
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
	return this.drain(ctx, outboxPageSize)
}

// DrainNow publishes what is waiting immediately, off the caller's goroutine.
//
// The sweep above is what GUARANTEES delivery; this only removes the wait for it. Until it existed
// the sweep was also the delivery path, so a customer standing at a vending machine waited for the
// next tick — up to a minute — between paying and the machine being told to dispense.
//
// Best-effort by design, and that is the whole point of the split: a failure here is not reported
// to anybody, because the row stays unpublished and the next sweep finds it. The worst case is
// exactly the behaviour before this existed.
//
// Callers reach this through services.DrainOutboxNow, which is where the rule about not draining
// from inside a transaction lives.
func (this *OutboxJobs) DrainNow(ctx corectx.Context) {
	if this.publisher == nil {
		return
	}

	// Detached before the goroutine starts: the caller answers its own request and its context is
	// cancelled, while this still has to read the outbox and reach the broker. CloneWithInner keeps
	// the scope, which is what the published message is stamped with.
	drainCtx := corectx.CloneWithInner(ctx, context.WithoutCancel(ctx.InnerContext()))

	go func() {
		defer func() {
			// A panic here would take the process down, and the write this is announcing has already
			// happened. Contained and logged; the sweep still has the row.
			if recovered := recover(); recovered != nil {
				this.logError("sales outbox: draining panicked", "", fmt.Errorf("%v", recovered))
			}
		}()

		if err := this.drain(drainCtx, drainPageSize); err != nil {
			this.logError("sales outbox: draining failed", "", err)
		}
	}()
}

// drain publishes one page of waiting events, marking each as it goes.
func (this *OutboxJobs) drain(ctx corectx.Context, limit int) error {
	if this.publisher == nil {
		// No broker adapter in this build; rows accumulate unpublished, which is the honest state.
		return nil
	}

	pending, err := services.UnpublishedEvents(ctx, limit)
	if err != nil {
		return err
	}

	// Detached, not blank. The publish must outlive whatever cancels the sweep's own context — the
	// producing request committed long ago and a retry may run hours later — but it must keep the
	// scope the sweep is running under, because that is what the publisher stamps on the message so
	// the consumer knows whose data the event is about. A context.Background() here published every
	// event unscoped, and the consumer then failed on its first query.
	publishCtx := corectx.CloneWithInner(ctx, context.WithoutCancel(ctx.InnerContext()))

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
// map, and the id is what an operator greps the outbox for. An empty id is a failure of the drain
// itself rather than of one event, and is left off rather than logged as "(event )".
func (this *OutboxJobs) logError(message, eventId string, err error) {
	if this.logger == nil {
		return
	}
	if eventId == "" {
		this.logger.Error(message, err)
		return
	}
	this.logger.Error(message+" (event "+eventId+")", err)
}
