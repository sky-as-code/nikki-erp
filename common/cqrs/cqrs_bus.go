package cqrs

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/common/logging"
	ft "github.com/sky-as-code/nikki-erp/common/util/fault"
)

const MetaCorrelationId = "correlation_id"
const MetaRequestTopic = "request_topic"
const MetaReplyTopic = "reply_topic"
const DefaultQueryTimeoutSecs = 50 * time.Second

func goChannelPubSub(logger logging.LoggerService) *gochannel.GoChannel {
	slogger := logging.Logger().InnerLogger().(*slog.Logger)

	watermill.NewSlogLoggerWithLevelMapping(slogger, map[slog.Level]slog.Level{
		// Watermill does not have a trace level, so we map it to warn,
		// so that we will call watermillLogger().Trace() to print warnings.
		watermill.LevelTrace: slog.LevelWarn,
	})

	return gochannel.NewGoChannel(gochannel.Config{}, watermill.NewSlogLogger(slogger))
}

type CqrsBusConfig struct {
	Logger logging.LoggerService

	// MaxTimeout is the maximum time to wait for a response.
	// Each Send() invocation can accept a context.WithTimeout with shorter wait time.
	//
	// This option is not required. Default is 50 seconds.
	MaxTimeout time.Duration
}

func (this CqrsBusConfig) setDefaults() {
	if this.MaxTimeout == 0 {
		this.MaxTimeout = DefaultQueryTimeoutSecs
	}
}

func (this CqrsBusConfig) Validate() error {
	var err error

	if this.Logger == nil {
		err = stdErrors.Join(err, errors.New("missing Logger"))
	}

	return err
}

func NewWatermillCqrsBus(config CqrsBusConfig) (*WatermillCqrsBus, error) {
	config.setDefaults()
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	pubSub := goChannelPubSub(config.Logger)
	marshaler := cqrs.ProtoMarshaler{
		GenerateName: cqrs.NamedStruct(cqrs.StructName),
	}

	return &WatermillCqrsBus{
		config:     config,
		publisher:  pubSub,
		subscriber: pubSub,
		marshaler:  marshaler,
	}, nil
}

type WatermillCqrsBus struct {
	config     CqrsBusConfig
	publisher  message.Publisher
	subscriber message.Subscriber
	marshaler  cqrs.CommandEventMarshaler
}

// Verify WaterMillCqrsBus implements CqrsBus interface
var _ CqrsBus = (*WatermillCqrsBus)(nil)

// SubscribeRequests registers multiple handlers under a single context, if the context is cancelled,
// those handlers' subscriptions will be cancelled.
func (this *WatermillCqrsBus) SubscribeRequests(ctx context.Context, handlers ...RequestHandler) (err error) {
	for _, handler := range handlers {
		err = stdErrors.Join(err, this.subscribeReq(ctx, handler))
	}
	return err
}

func (this *WatermillCqrsBus) subscribeReq(ctx context.Context, handler RequestHandler) (err error) {
	ctx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to subscribe to requests")
			cancelSubscription()
		}
	}()

	sampleRequest := handler.NewRequest()
	if !isPointer(sampleRequest) {
		return errors.New("handler.NewRequest() must return pointer to the new request instance")
	}

	topicName := genRequestTopic(handler.NewRequest().Type().String())
	msgChan, err := this.subscriber.Subscribe(ctx, topicName)
	ft.PanicOnErr(err)

	go func() {
		defer cancelSubscription()

		for {
			select {
			case msg := <-msgChan:
				var reply Reply
				request := handler.NewRequest()
				reqPacket, err := newIncomingRequestPacket(msg, this.marshaler, request)
				if err != nil {
					c, _ := context.WithTimeout(context.Background(), this.config.MaxTimeout)
					r, err := handler.Handle(c, reqPacket)
					if err != nil {
						reply.Error = err
					} else {
						reply = *r
					}
					replyPacket := newReplyPacket(reqPacket.correlationId, reply, this.marshaler)
					err = this.publisher.Publish(reqPacket.replyTopic, replyPacket.message)
					if err != nil {
						this.config.Logger.Error(
							fmt.Sprintf("failed to publish reply to topic %s", reqPacket.replyTopic),
							err,
						)
					}
				} else {
					this.config.Logger.Error(
						fmt.Sprintf("failed to parse request from topic %s", topicName),
						err,
					)
				}
			case <-ctx.Done():
				err = ctx.Err()
				return
			}
		}
	}()

	return nil
}

func (this *WatermillCqrsBus) RequestNoReply(ctx context.Context, request Request) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to send request")
		}
	}()

	packet, err := this.newRequestPacket(ctx, request)
	ft.PanicOnErr(err)

	err = this.publisher.Publish(packet.requestTopic, packet.message)
	ft.PanicOnErr(err)

	return nil
}

func (this *WatermillCqrsBus) Request(ctx context.Context, request Request) (_ <-chan Reply, err error) {
	ctx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to send request")
			cancelSubscription()
		}
	}()

	packet, err := this.newRequestPacket(ctx, request)
	ft.PanicOnErr(err)

	replyChan := make(chan Reply, 1)
	err = this.subscribeReply(ctx, packet, replyChan, cancelSubscription)
	ft.PanicOnErr(err)

	err = this.publisher.Publish(packet.requestTopic, packet.message)
	ft.PanicOnErr(err)

	return replyChan, nil
}

func (this *WatermillCqrsBus) subscribeReply(ctx context.Context, packet *RequestPacket, replyChan chan Reply, cancelSubscription context.CancelFunc) error {
	msgChan, err := this.subscriber.Subscribe(ctx, packet.replyTopic)
	if err != nil {
		return err
	}

	go func() {
		defer close(replyChan)
		defer cancelSubscription()

		select {
		case msg := <-msgChan:
			var reply Reply
			err = this.marshaler.Unmarshal(msg, &reply)
			if err != nil {
				replyChan <- Reply{Error: err}
				return
			}
			replyChan <- reply
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(this.config.MaxTimeout):
			err = errors.Errorf("timeout waiting for reply")
		}
	}()

	return nil
}

func (this *WatermillCqrsBus) newRequestPacket(ctx context.Context, request Request) (packet *RequestPacket, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to create request packet")
		}
	}()
	packet = newOutgoingRequestPacket(request, this.marshaler)
	packet.message.SetContext(ctx)

	return packet, nil
}

func genRequestTopic(requestType string) string {
	return "cqrs:" + requestType
}

func genReplyTopic(requestTopic string, correlationId string) string {
	return fmt.Sprintf("%s:reply:%s", requestTopic, correlationId)
}

func newOutgoingRequestPacket(request Request, marshaler cqrs.CommandEventMarshaler) *RequestPacket {
	msg, err := marshaler.Marshal(request)
	ft.PanicOnErr(err)

	packet := &RequestPacket{
		message: msg,
	}

	newUlid, err := ulid.New()
	ft.PanicOnErr(err)

	packet.correlationId = newUlid.String()
	requestType := marshaler.Name(request)
	packet.requestTopic = genRequestTopic(requestType)
	msg.Metadata.Set(MetaCorrelationId, packet.correlationId)
	msg.Metadata.Set(MetaRequestTopic, packet.requestTopic)
	msg.Metadata.Set(MetaReplyTopic, genReplyTopic(packet.requestTopic, packet.correlationId))

	return packet
}

func newIncomingRequestPacket(
	msg *message.Message,
	marshaler cqrs.CommandEventMarshaler,
	request Request,
) (*RequestPacket, error) {
	packet := &RequestPacket{
		message: msg,
	}

	err := marshaler.Unmarshal(msg, request)
	if err != nil {
		return nil, err
	}

	packet.request = request
	packet.correlationId = msg.Metadata.Get(MetaCorrelationId)
	packet.requestTopic = msg.Metadata.Get(MetaRequestTopic)
	packet.replyTopic = msg.Metadata.Get(MetaReplyTopic)

	return packet, nil
}

func newReplyPacket(correlationId string, reply Reply, marshaler cqrs.CommandEventMarshaler) *ReplyPacket {
	msg, err := marshaler.Marshal(reply)
	ft.PanicOnErr(err)

	packet := &ReplyPacket{
		message: msg,
	}

	packet.correlationId = correlationId
	msg.Metadata.Set(MetaCorrelationId, packet.correlationId)

	return packet
}

type RequestPacket struct {
	correlationId string
	requestTopic  string
	replyTopic    string
	message       *message.Message
	request       Request
}

func (this RequestPacket) CorrelationId() string {
	return this.correlationId
}

func (this RequestPacket) Request() Request {
	return this.request
}

type ReplyPacket struct {
	correlationId string
	message       *message.Message
	reply         Reply
}

func (this ReplyPacket) CorrelationId() string {
	return this.correlationId
}

func (this ReplyPacket) Reply() Reply {
	return this.reply
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

type RequestHandler interface {
	Handle(ctx context.Context, packet *RequestPacket) (*Reply, error)

	// Type returns the type of request handled by this handler
	// Type() RequestType

	// NewRequest returns a new instance of the request type handled by this handler
	NewRequest() Request
}

type RequestType struct {
	Module    string
	Submodule string
	Action    string
}

func (this RequestType) String() string {
	return this.Module + "_" + this.Submodule + "." + this.Action
}

type Request interface {
	Name() string
	Type() RequestType
}

func isPointer(v any) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}

// type NonPointerError struct {
// 	Type reflect.Type
// }

// func (e NonPointerError) Error() string {
// 	return "non-pointer command: " + e.Type.String() + ", handler.NewCommand() should return pointer to the command"
// }
