package cqrs

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const MetaCorrelationId = "correlation_id"
const MetaRequestTopic = "request_topic"
const MetaReplyTopic = "reply_topic"
const MetaNoReply = "no_reply"
const DefaultQueryTimeoutSecs = "50"

type CqrsBusParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

func NewWatermillCqrsBus(params CqrsBusParams) (CqrsBus, error) {
	pubSub := goChannelPubSub(params.Logger)
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.NamedStruct(cqrs.StructName),
	}
	maxTimeoutSec := params.Config.GetInt(c.CqrsRequestTimeoutSecs, DefaultQueryTimeoutSecs)

	return &WatermillCqrsBus{
		logger:     params.Logger,
		publisher:  pubSub,
		subscriber: pubSub,
		marshaler:  marshaler,
		maxTimeout: time.Duration(maxTimeoutSec) * time.Second,
	}, nil
}

func goChannelPubSub(logger logging.LoggerService) *gochannel.GoChannel {
	slogger := logger.InnerLogger().(*slog.Logger)

	watermill.NewSlogLoggerWithLevelMapping(slogger, map[slog.Level]slog.Level{
		// Watermill does not have a trace level, so we map it to warn,
		// so that we will call watermillLogger().Trace() to print warnings.
		watermill.LevelTrace: slog.LevelWarn,
	})

	return gochannel.NewGoChannel(gochannel.Config{}, watermill.NewSlogLogger(slogger))
}

type WatermillCqrsBus struct {
	logger        logging.LoggerService
	marshaler     cqrs.CommandEventMarshaler
	publisher     message.Publisher
	subscriber    message.Subscriber
	subscriptions sync.Map

	maxTimeout time.Duration
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
	defer func() {
		err = ft.RecoverPanicf(recover(), "failed to subscribe with handler %s", structName(handler))
	}()

	sampleRequest := handler.NewRequest().(Request)

	requestType := sampleRequest.Type().String()

	if _, existing := this.subscriptions.Load(requestType); existing {
		return errors.Errorf("request type %s is already handled", requestType)
	}

	this.subscriptions.Store(requestType, handler)
	ctx, cancelContext := context.WithCancel(ctx)

	cancelSubscription := func() {
		cancelContext()
		this.subscriptions.Delete(requestType)
	}

	defer func() {
		if err != nil {
			cancelSubscription()
		}
	}()

	topicName := genRequestTopic(requestType)
	msgChan, err := this.subscriber.Subscribe(ctx, topicName)
	ft.PanicOnErr(err)

	go func() {
		defer cancelSubscription()

		for {
			select {
			case msg := <-msgChan:
				request := handler.NewRequest()
				reply := handler.NewReply()
				reqPacket, err := newIncomingRequestPacket(msg, this.marshaler, request)
				msg.Ack()
				if err != nil {
					this.logger.Error(
						fmt.Sprintf("failed to parse request from topic %s", topicName),
						err,
					)
					continue
				}
				c, _ := context.WithTimeout(context.Background(), this.maxTimeout)
				r, err := handler.Handle(c, reqPacket)
				if err != nil {
					this.logger.Error(
						fmt.Sprintf("error occured from topic %s", topicName),
						err,
					)
					reply.Error = util.ToPtr(err.Error())
				} else {
					reply = *r
				}
				replyPacket := newReplyPacket(reqPacket.correlationId, &reply, this.marshaler)
				err = this.publisher.Publish(reqPacket.replyTopic, replyPacket.message)
				if err != nil {
					this.logger.Error(
						fmt.Sprintf("failed to publish reply to topic %s", reqPacket.replyTopic),
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
		err = ft.RecoverPanic(recover(), "failed to send request")
	}()

	packet, err := this.newRequestPacket(ctx, request)
	ft.PanicOnErr(err)
	packet.message.Metadata.Set(MetaNoReply, "true")

	err = this.publisher.Publish(packet.requestTopic, packet.message)
	ft.PanicOnErr(err)

	return nil
}

func (this *WatermillCqrsBus) Request(ctx context.Context, request Request, result any) (err error) {
	ctx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to send request of type %s", request.Type().String()); e != nil {
			err = e
			cancelSubscription()
		}
	}()

	packet, err := this.newRequestPacket(ctx, request)
	ft.PanicOnErr(err)

	replyChan, errChan := this.subscribeReply(ctx, packet, result, cancelSubscription)
	ft.PanicOnErr(err)

	err = this.publisher.Publish(packet.requestTopic, packet.message)
	ft.PanicOnErr(err)

	select {
	case reply := <-replyChan:
		if reply.Error != nil {
			return errors.New(*reply.Error)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func (this *WatermillCqrsBus) subscribeReply(ctx context.Context, packet *RequestPacket[Request], result any, cancelSubscription context.CancelFunc) (<-chan *Reply[any], <-chan error) {
	replyChan := make(chan *Reply[any])
	errChan := make(chan error)

	handleErr := func() {
		if r := recover(); r != nil {
			err := errors.Wrap(r.(error), fmt.Sprintf("failed to subscribe for reply from topic %s", packet.replyTopic))
			errChan <- err
			close(errChan)
			close(replyChan)
		}
	}

	defer handleErr()

	msgChan, err := this.subscriber.Subscribe(ctx, packet.replyTopic)
	if err != nil {
		errChan <- err
		return replyChan, errChan
	}

	go func() {
		defer cancelSubscription()
		defer handleErr()

		select {
		case msg := <-msgChan:
			msg.Ack()
			reply := &Reply[any]{
				Result: result,
			}
			err = this.marshaler.Unmarshal(msg, reply)
			if err == nil {
				replyChan <- reply
				close(replyChan)
				close(errChan)
				return
			}
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(this.maxTimeout):
			err = errors.Errorf("timeout for request %s (%s)", packet.correlationId, packet.requestTopic)
		}

		// If we reach here, it means we have an error,
		// close error channel first to follow the failure path
		errChan <- err
		close(errChan)
		close(replyChan)
	}()

	return replyChan, errChan
}

func (this *WatermillCqrsBus) newRequestPacket(ctx context.Context, request Request) (packet *RequestPacket[Request], err error) {
	defer func() {
		err = ft.RecoverPanicf(recover(), "failed to create request packet for %s", request.Type().String())
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

func newOutgoingRequestPacket(request Request, marshaler cqrs.CommandEventMarshaler) *RequestPacket[Request] {
	msg, err := marshaler.Marshal(&request)
	ft.PanicOnErr(err)

	packet := &RequestPacket[Request]{
		message: msg,
	}

	newUlid, err := ulid.New()
	ft.PanicOnErr(err)

	packet.correlationId = newUlid.String()
	requestType := request.Type().String()
	packet.requestTopic = genRequestTopic(requestType)
	packet.replyTopic = genReplyTopic(packet.requestTopic, packet.correlationId)
	msg.Metadata.Set(MetaCorrelationId, packet.correlationId)
	msg.Metadata.Set(MetaRequestTopic, packet.requestTopic)
	msg.Metadata.Set(MetaReplyTopic, packet.replyTopic)

	return packet
}

func newIncomingRequestPacket(
	msg *message.Message,
	marshaler cqrs.CommandEventMarshaler,
	request any,
) (*RequestPacket[Request], error) {
	packet := &RequestPacket[Request]{
		message: msg,
	}

	err := marshaler.Unmarshal(msg, request)
	if err != nil {
		return nil, err
	}

	packet.request = request.(Request)
	packet.correlationId = msg.Metadata.Get(MetaCorrelationId)
	packet.requestTopic = msg.Metadata.Get(MetaRequestTopic)
	packet.replyTopic = msg.Metadata.Get(MetaReplyTopic)

	return packet, nil
}

func newReplyPacket(correlationId string, reply *Reply[any], marshaler cqrs.CommandEventMarshaler) *ReplyPacket[any] {
	msg, err := marshaler.Marshal(reply)
	ft.PanicOnErr(err)

	packet := &ReplyPacket[any]{
		message: msg,
	}

	packet.correlationId = correlationId
	msg.Metadata.Set(MetaCorrelationId, packet.correlationId)

	return packet
}

// func isPointer(v any) bool {
// 	return reflect.ValueOf(v).Kind() == reflect.Ptr
// }

func structName(v any) string {
	return reflect.ValueOf(v).Kind().String()
}

func NewHandler[TReq Request, TResult any](
	handleFunc func(ctx context.Context, packet *RequestPacket[TReq]) (*Reply[TResult], error),
) RequestHandler {
	return &genericRequestHandler[TReq, TResult]{
		handleFunc: handleFunc,
	}
}

type genericRequestHandler[TReq Request, TResult any] struct {
	handleFunc func(ctx context.Context, packet *RequestPacket[TReq]) (*Reply[TResult], error)
}

func (c genericRequestHandler[TReq, TResult]) NewRequest() any {
	var val TReq
	return &val
}

func (c genericRequestHandler[TReq, TResult]) NewReply() Reply[any] {
	var result TResult
	var val Reply[any]
	val.Result = result
	return val
}

func (c genericRequestHandler[TReq, TResult]) Handle(ctx context.Context, packet *RequestPacket[Request]) (reply *Reply[any], err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle request"); e != nil {
			err = e
		}
	}()

	var req any = packet.request
	typedReq := req.(*TReq)
	packet.request = *typedReq
	typedPacket := &RequestPacket[TReq]{
		correlationId: packet.correlationId,
		requestTopic:  packet.requestTopic,
		replyTopic:    packet.replyTopic,
		message:       packet.message,
		request:       packet.request.(TReq),
	}
	typedReply, err := c.handleFunc(ctx, typedPacket)
	if err != nil {
		return nil, err
	}

	reply = &Reply[any]{
		Result: typedReply.Result,
		Error:  typedReply.Error,
	}
	return reply, err
}
