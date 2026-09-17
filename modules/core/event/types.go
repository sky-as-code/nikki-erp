package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type EventBus interface {
	PublishRequest(ctx context.Context, request EventRequest) (err error)
	PublishRequestWaitReply(ctx context.Context, request EventRequest, DataReply any) (reply *Reply[any], err error)
	PublishReply(ctx context.Context, request EventRequest, reply *Reply[any]) (err error)
	SubscribeEvent(ctx context.Context, request EventRequest, prototype any) (deliveryChan chan EventDelivery, err error)
	Close() error
}

// EventDelivery is one message as its subscriber receives it: the decoded event, plus the scope the
// publisher was executing under.
//
// # Why the scope rides along
//
// A subscriber consumes on a goroutine registered at boot, whose context descends from
// context.Background(). It therefore names no tenant and no organization, while every read below it
// needs both — and in a multi-tenant build the repository PANICS rather than querying across
// tenants. Before this envelope existed each pub/sub pair solved that for itself, by adding a
// TenantId field to its own event and re-attaching it by hand on the other side. Three pairs had
// done so, each with its own copy of the "tenant_id" key; a fourth that forgot would only find out
// when a repository panicked under real traffic.
//
// The scope now travels the way it already does across the CQRS bus — in the message's metadata,
// carried by the bus itself (see MetaDomainConstraints) — so an event's payload describes only what
// happened, and a new pub/sub pair inherits the scoping without doing anything.
type EventDelivery struct {
	// Payload is the decoded event: a pointer to a value of the prototype's type, freshly allocated
	// for this message, so a handler may keep it.
	Payload any

	// Constraints is what the publisher's context was scoped by, the tenant above all. Nil when it
	// was scoped by nothing, which is every publish in a single-tenant build.
	Constraints dmodel.DynamicFields
}

// ScopeContext re-establishes the publisher's scope on the context a handler will be given.
//
// A no-op when the publisher carried no constraints, which leaves a single-tenant build's contexts
// exactly as they were rather than scoping them to an empty map.
func (this EventDelivery) ScopeContext(ctx corectx.Context) {
	if len(this.Constraints) > 0 {
		ctx.SetDomainConstraints(this.Constraints)
	}
}

type EventRequest struct {
	correlationId string
	eventTopic    string
	replyTopic    string
	message       *message.Message
}

func NewEventRequest(correlationId, eventTopic, replyTopic string, message *message.Message) *EventRequest {
	return &EventRequest{
		correlationId: correlationId,
		eventTopic:    eventTopic,
		replyTopic:    replyTopic,
		message:       message,
	}
}

type Reply[TResult any] struct {
	Result TResult `json:"result"`
	Error  *string `json:"error"`
}

type EventReply[TResult any] struct {
	correlationId string
	message       *message.Message
	reply         Reply[TResult]
}

func (packet EventReply[TResult]) CorrelationId() string {
	return packet.correlationId
}

func (packet EventReply[TResult]) Reply() *Reply[TResult] {
	return &packet.reply
}
