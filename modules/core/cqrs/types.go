package cqrs

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type CqrsBus interface {
	SubscribeRequests(ctx context.Context, handlers ...RequestHandler) error
	RequestNoReply(ctx context.Context, request Request) error
	Request(ctx context.Context, request Request, result any) error
}

// Deprecated: Not used
// type Namer interface {
// 	Name() string
// }

type RequestType struct {
	Module    string `json:"module"`
	Submodule string `json:"submodule"`
	Action    string `json:"action"`
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
	Result TResult `json:"result"`
	Error  *string `json:"error"`
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
	NewRequest() any

	// NewReply returns a new instance of the reply type returned by this handler
	NewReply() Reply[any]
}
