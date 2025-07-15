package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type EventBus interface {
	PublishRequest(ctx context.Context, request EventRequest) (err error)
	PublishRequestWaitReply(ctx context.Context, request EventRequest, DataReply any) (reply *Reply[any], err error)
	PublishReply(ctx context.Context, request EventRequest, reply *Reply[any]) (err error)
	SubscribeRequest(ctx context.Context, request EventRequest, result any) (requestChan chan any, err error)
	Close() error
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
