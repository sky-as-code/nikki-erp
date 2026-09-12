// Package external binds Notification to the infrastructure and to other modules. It is the only
// place in the module that imports another module's interfaces.
package external

import (
	"context"
	"encoding/json"
	"fmt"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// PubSubBroker is the RealtimeNotificationBroker on top of the shared pub/sub transport.
//
// It deliberately adds no delivery guarantee of its own. A lost signal costs a client its realtime
// push, not its notification: the row is already committed, and the client reconciles by reading
// the inbox or by reconnecting with after_seq (BR-FS 18).
type PubSubBroker struct {
	publisher  pubsub.Publisher
	subscriber pubsub.Subcriber
	logger     logging.LoggerService
	topic      string
}

func NewPubSubBroker(
	publisher pubsub.Publisher,
	subscriber pubsub.Subcriber,
	cfg config.ConfigService,
	logger logging.LoggerService,
) it.RealtimeNotificationBroker {
	prefix := cfg.GetStr(modconstants.StreamTopicPrefix)
	return &PubSubBroker{
		publisher:  publisher,
		subscriber: subscriber,
		logger:     logger,
		// One topic for every instance rather than one per user: a per-user topic would mean a
		// subscribe and unsubscribe on the broker for every page load, and the fan-out filter is
		// a map lookup this instance does anyway.
		topic: fmt.Sprintf("%s.wakeup", prefix),
	}
}

func (this *PubSubBroker) Publish(ctx context.Context, signal it.WakeUpSignal) error {
	payload, err := json.Marshal(signal)
	if err != nil {
		return errors.Wrap(err, "marshal notification wake-up signal")
	}
	return this.publisher.Publish(ctx, this.topic, payload)
}

func (this *PubSubBroker) Subscribe(ctx context.Context) (<-chan it.WakeUpSignal, error) {
	raw, err := this.subscriber.Subscribe(ctx, this.topic)
	if err != nil {
		return nil, errors.Wrap(err, "subscribe to notification wake-up topic")
	}

	signals := make(chan it.WakeUpSignal, 256)
	go this.decodeLoop(ctx, raw, signals)
	return signals, nil
}

// decodeLoop turns broker bytes into signals, dropping anything it cannot read.
//
// A malformed message is logged and skipped rather than closing the subscription: the broker is
// shared, and one bad publisher must not take realtime delivery down for every user on this
// instance.
func (this *PubSubBroker) decodeLoop(
	ctx context.Context, raw <-chan []byte, signals chan<- it.WakeUpSignal,
) {
	defer close(signals)

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-raw:
			if !ok {
				return
			}
			var signal it.WakeUpSignal
			if err := json.Unmarshal(message, &signal); err != nil {
				this.logger.Error("notification wake-up signal is not readable", err)
				continue
			}
			select {
			case signals <- signal:
			case <-ctx.Done():
				return
			}
		}
	}
}
