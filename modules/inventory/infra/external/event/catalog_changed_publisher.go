package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreEvent "github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itEvent "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/event"
)

const publishAsyncTimeout = time.Minute

type CatalogChangedEventPublisherImpl struct {
	bus    coreEvent.EventBus
	logger logging.LoggerService
	topic  string
}

func NewCatalogChangedEventPublisher(
	bus coreEvent.EventBus, cfg config.ConfigService, logger logging.LoggerService,
) itEvent.CatalogChangedEventPublisher {
	return &CatalogChangedEventPublisherImpl{
		bus:    bus,
		logger: logger,
		topic:  cfg.GetStr(constants.CatalogChangedEventTopic, constants.DefaultCatalogChangedEventTopic),
	}
}

func (this *CatalogChangedEventPublisherImpl) Publish(
	ctx corectx.Context, eventType itEvent.CatalogChangedType, orgId string, ids ...string,
) error {
	if orgId == "" {
		return nil
	}

	filtered := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			filtered = append(filtered, id)
		}
	}

	event := itEvent.CatalogChangedEvent{Type: eventType, OrgId: orgId, Ids: filtered}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req := coreEvent.NewEventRequest("", this.topic, "", &message.Message{Payload: body})

	return this.bus.PublishRequest(ctx.InnerContext(), *req)
}

func (this *CatalogChangedEventPublisherImpl) PublishAsync(
	ctx corectx.Context, eventType itEvent.CatalogChangedType, orgId string, ids ...string,
) {
	go func() {
		bgCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx.InnerContext()), publishAsyncTimeout)
		defer cancel()

		if err := this.Publish(corectx.NewRequestContext(bgCtx), eventType, orgId, ids...); err != nil {
			this.logger.Warn("catalog changed event publish failed",
				logging.Attr{"error": err.Error(), "resource_type": eventType, "org_id": orgId, "ids": ids})
		}
	}()
}
