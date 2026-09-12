package services

import (
	"context"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// StreamDispatcher turns wake-up signals into events on the streams this instance holds.
//
// It is the piece that makes several backend instances behave as one: the notification may be
// created on instance B while the person reading it is connected to instance A, and without this
// the event would simply never arrive (BR 17, BR-FS 18).
type StreamDispatcher struct {
	broker     it.RealtimeNotificationBroker
	registry   it.StreamRegistry
	recipients it.RecipientRepository
	logger     logging.LoggerService
}

func NewStreamDispatcher(
	broker it.RealtimeNotificationBroker,
	registry it.StreamRegistry,
	recipients it.RecipientRepository,
	logger logging.LoggerService,
) *StreamDispatcher {
	return &StreamDispatcher{
		broker:     broker,
		registry:   registry,
		recipients: recipients,
		logger:     logger,
	}
}

// Run consumes wake-up signals until the context is cancelled.
func (this *StreamDispatcher) Run(ctx context.Context) error {
	signals, err := this.broker.Subscribe(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case signal, open := <-signals:
				if !open {
					return
				}
				this.dispatch(ctx, signal)
			}
		}
	}()
	return nil
}

// dispatch reads what the signal points at and writes it to the local streams.
//
// The read is the point: the signal says only THAT something arrived, and the notification itself
// comes from the database (BR 17, BR-FS 18). Trusting the signal's payload would make a lost or
// reordered message able to put wrong content on a person's screen.
func (this *StreamDispatcher) dispatch(ctx context.Context, signal it.WakeUpSignal) {
	// Nobody here is connected to this person, so there is nothing to read for. Every instance
	// receives every signal, and on all but one of them this is the path taken.
	if !this.registry.HasStreams(signal.OrgId, signal.UserId) {
		return
	}

	afterSeq := signal.Seq - 1
	requestCtx := corectx.NewRequestContext(ctx)

	items, err := this.recipients.Replay(requestCtx, signal.OrgId, signal.UserId, &afterSeq, dispatchPageSize)
	if err != nil {
		// The notification is durable and the client reconciles through the inbox or an
		// after_seq reconnect, so a failed read costs a realtime push and nothing more.
		this.logger.Error("notification wake-up could not be read back", err)
		return
	}

	for _, item := range items {
		seq := item.StreamSeq
		this.registry.Dispatch(signal.OrgId, signal.UserId, it.StreamEvent{
			Seq:  &seq,
			Type: it.StreamEventNotificationCreated,
			Data: item,
		})
	}
}

// dispatchPageSize bounds one wake-up read. A signal points at one notification; the margin is for
// the several that can arrive between the signal being published and this reading it.
const dispatchPageSize = 20
