package cqrs

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"go.bryk.io/pkg/ulid"
)

const MetaCorrelationId = "correlation_id"
const MetaResponseTopic = "response_topic"

type QueryEventMarshaler = cqrs.CommandEventMarshaler

type QueryPacket struct {
	correlationId string
	query         any
}

func NewQueryPacket(query any) QueryPacket {
	corId, err := ulid.New()
	if err != nil {
		panic(err)
	}
	return QueryPacket{
		correlationId: corId.String(),
		query:         query,
	}
}

func (this QueryPacket) CorrelationId() string {
	return this.correlationId
}

func (this QueryPacket) GetQuery() any {
	return this.query
}

type Reply struct {
	// HandlerResult contains the handler result.
	// It's preset only when NewCommandHandlerWithResult is used. If NewCommandHandler is used, HandlerResult is empty.
	//
	// Result is sent even if the handler returns an error.
	Result any

	// Error contains the error returned by the command handler or the Backend when handling notification fails.
	// Handling the notification can fail, for example, when unmarshaling the message or if there's a timeout.
	// If listening for a reply times out or the context is canceled, the Error is ReplyTimeoutError.
	//
	// If an error from the handler is returned, CommandHandlerError is returned.
	// If processing was successful, Error is nil.
	Error error
}
