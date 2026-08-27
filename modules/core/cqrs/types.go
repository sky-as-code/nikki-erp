package cqrs

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type CqrsBus interface {
	SubscribeRequests(ctx context.Context, handlers ...RequestHandler) error
	RequestNoReply(ctx context.Context, request Request) error
	Request(ctx context.Context, request Request, result any) error

	// IsRequestTypeRegistered reports whether a handler is currently subscribed for
	// requestType, formatted as RequestType.String() ("{module}_{submodule}.{action}").
	//
	// It answers "currently subscribed", not "ever registered": a subscription is removed
	// when its context is cancelled. That is the useful reading for a caller deciding
	// whether emitting a request would reach anybody.
	//
	// Callers must be aware of boot ordering. Handlers subscribe during their module's
	// Init(), so this reports false for a module that has not been initialized yet. It is
	// safe from anything that runs after startup completes - a REST handler, or a module's
	// OnAppStarted - and unreliable from another module's Init().
	IsRequestTypeRegistered(requestType string) bool
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
	CqrsRequestType() RequestType
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
