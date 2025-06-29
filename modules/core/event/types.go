package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type EventBus interface {
	PublishNoReply(ctx context.Context, packet *EventPacket) error
	PublishWaitReply(ctx context.Context, packet *EventPacket, result any) error
	PublishEvent(ctx context.Context, topic string, eventData any) error
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	SubscribeReply(ctx context.Context, replyTopic string, handler ReplyHandler) error
	PublishReply(ctx context.Context, replyTopic string, reply any, correlationId string) error
	Close() error
}
type EventPacket struct {
	correlationId string
	eventTopic    string
	replyTopic    string
	message       *message.Message
}

func (packet EventPacket) CorrelationId() string {
	return packet.correlationId
}

type Reply[TResult any] struct {
	Result TResult `json:"result"`
	Error  *string `json:"error"`
}

type ReplyPacket[TResult any] struct {
	correlationId string
	reply         Reply[TResult]
}

func (packet ReplyPacket[TResult]) CorrelationId() string {
	return packet.correlationId
}

func (packet ReplyPacket[TResult]) Reply() *Reply[TResult] {
	return &packet.reply
}

type EventHandler interface {
	Handle(ctx context.Context, packet *EventPacket) error
	NewEvent() any
}

type ReplyHandler interface {
	Handle(ctx context.Context, packet *ReplyPacket[any]) error
}
