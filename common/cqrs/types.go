package cqrs

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type CqrsBus interface {
	SubscribeRequests(ctx context.Context, handlers ...RequestHandler) (err error)
	RequestNoReply(ctx context.Context, request Request) (err error)
	Request(ctx context.Context, request Request) (_ <-chan Reply[any], err error)
}

// Deprecated: Not used
// type Namer interface {
// 	Name() string
// }

type RequestType struct {
	Module    string
	Submodule string
	Action    string
}

func (this RequestType) String() string {
	return this.Module + "_" + this.Submodule + "." + this.Action
}

type Request interface {
	Type() RequestType
}

type RequestPacket[TReq Request] struct {
	correlationId string
	requestTopic  string
	replyTopic    string
	message       *message.Message
	request       TReq
}

func (this RequestPacket[TReq]) CorrelationId() string {
	return this.correlationId
}

func (this RequestPacket[TReq]) Request() *TReq {
	return &this.request
}

type Reply[TResult any] struct {
	Result TResult
	Error  error
}

type ReplyPacket[TResult any] struct {
	correlationId string
	message       *message.Message
	reply         Reply[TResult]
}

func (this ReplyPacket[TResult]) CorrelationId() string {
	return this.correlationId
}

func (this ReplyPacket[TResult]) Reply() *Reply[TResult] {
	return &this.reply
}

type RequestHandler interface {
	Handle(ctx context.Context, packet *RequestPacket[Request]) (*Reply[any], error)

	// Type returns the type of request handled by this handler
	// Type() RequestType

	// NewRequest returns a new instance of the request type handled by this handler
	NewRequest() Request

	// NewReply returns a new instance of the reply type returned by this handler
	NewReply() Reply[any]
}
