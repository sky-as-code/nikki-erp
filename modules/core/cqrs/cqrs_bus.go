package cqrs

import (
	"context"
	stdErrors "errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/message/transports"
)

const MetaCorrelationId = "correlation_id"
const MetaDomainConstraints = "domain_constraints"
const MetaRequestTopic = "request_topic"
const MetaReplyTopic = "reply_topic"
const MetaNoReply = "no_reply"
const DefaultQueryTimeoutSecs = "50"

// Metadata keys a scheduled job attaches to the command it triggers, so a handler can make its
// side effect idempotent by the same means whether it was invoked over HTTP or over the bus.
//
// They are declared here rather than in the scheduler because they name a convention on the
// envelope: a handler reading them must not have to import the module that happens to set them.
const MetaIdempotencyKey = "idempotency_key"
const MetaSchedulerJobId = "scheduler_job_id"
const MetaSchedulerExecutionId = "scheduler_execution_id"
const MetaSchedulerAttempt = "scheduler_attempt"

// reservedMetaKeys are the envelope keys the bus owns. Caller-supplied metadata may not use
// them, so that a caller cannot forge a correlation id or redirect a reply.
var reservedMetaKeys = map[string]bool{
	MetaCorrelationId:     true,
	MetaDomainConstraints: true,
	MetaRequestTopic:      true,
	MetaReplyTopic:        true,
	MetaNoReply:           true,
}

type cqrsMetadataKey struct{}

// WithMetadata returns a context carrying md, to be copied onto the envelope of any request
// sent with that context.
//
// It exists so a caller can pass information alongside a request that is not part of the
// request's own type - an idempotency key identifying the work that triggered it, for
// instance - without every such field having to become a member of the command struct.
//
// Reserved keys in md are ignored rather than rejected, because the caller has no way to
// know which names the bus reserves and silently dropping one is safer than either honoring
// it or failing the send.
func WithMetadata(ctx context.Context, md map[string]string) context.Context {
	if len(md) == 0 {
		return ctx
	}
	merged := map[string]string{}
	if existing, ok := ctx.Value(cqrsMetadataKey{}).(map[string]string); ok {
		for key, val := range existing {
			merged[key] = val
		}
	}
	for key, val := range md {
		if !reservedMetaKeys[key] {
			merged[key] = val
		}
	}
	return context.WithValue(ctx, cqrsMetadataKey{}, merged)
}

// MetadataFrom returns the caller-supplied metadata carried by ctx, or nil. Handlers use it
// to read what the sender attached with WithMetadata.
func MetadataFrom(ctx context.Context) map[string]string {
	md, _ := ctx.Value(cqrsMetadataKey{}).(map[string]string)
	return md
}

type CqrsBusParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService

	Transport *transports.MessageTransport `name:"go-channel"`
}

func NewWatermillCqrsBus(params CqrsBusParams) (CqrsBus, error) {
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.NamedStruct(cqrs.StructName),
	}
	maxTimeoutSec := params.Config.GetInt(c.CqrsRequestTimeoutSecs, DefaultQueryTimeoutSecs)

	return &WatermillCqrsBus{
		logger:     params.Logger,
		publisher:  params.Transport.Publisher,
		subscriber: params.Transport.Subscriber,
		marshaler:  marshaler,
		maxTimeout: time.Duration(maxTimeoutSec) * time.Second,
	}, nil
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
		err = ft.RecoverPanicFailedTo(recover(), "subscribe with handler "+structName(handler))
	}()

	sampleRequest := handler.NewRequest().(Request)

	requestType := sampleRequest.CqrsRequestType().String()

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
				reqCtx := this.createIncomingContext(c, msg)
				reply, err = this.handleRequest(reqCtx, handler, reqPacket, topicName)
				this.sendResponse(reqPacket, &reply)
				// r, err := handler.Handle(reqCtx, reqPacket)
				// if err != nil {
				// 	this.logger.Error(
				// 		fmt.Sprintf("error occured from topic %s", topicName),
				// 		err,
				// 	)
				// 	reply.Error = util.ToPtr(err.Error())
				// } else {
				// 	reply = *r
				// }
				// replyPacket := newReplyPacket(reqPacket.correlationId, &reply, this.marshaler)
				// err = this.publisher.Publish(reqPacket.replyTopic, replyPacket.message)
				// if err != nil {
				// 	this.logger.Error(
				// 		fmt.Sprintf("failed to publish reply to topic %s", reqPacket.replyTopic),
				// 		err,
				// 	)
				// }
			case <-ctx.Done():
				err = ctx.Err()
				return
			}
		}
	}()

	return nil
}

// IsRequestTypeRegistered implements CqrsBus by consulting the same subscriptions map that
// subscribeReq writes to and cancelSubscription deletes from, so it cannot drift from the
// set of handlers actually listening.
func (this *WatermillCqrsBus) IsRequestTypeRegistered(requestType string) bool {
	_, exists := this.subscriptions.Load(requestType)
	return exists
}

func (this *WatermillCqrsBus) createIncomingContext(ctx context.Context, msg *message.Message) context.Context {
	// Re-attach the sender's metadata before building the request context, so a handler can
	// read it with MetadataFrom exactly as the sender wrote it with WithMetadata. Reserved
	// keys are left out: they describe this envelope's delivery, not the caller's intent.
	callerMeta := map[string]string{}
	for key, vals := range msg.Metadata {
		if !reservedMetaKeys[key] {
			callerMeta[key] = vals
		}
	}
	if len(callerMeta) > 0 {
		ctx = context.WithValue(ctx, cqrsMetadataKey{}, callerMeta)
	}

	reqCtx := corectx.NewRequestContext(ctx)
	domConstrStr := msg.Metadata.Get(MetaDomainConstraints)
	if domConstrStr != "" {
		domConstr := dmodel.DynamicFields{}
		err := json.Unmarshal([]byte(domConstrStr), &domConstr)
		ft.PanicOnErr(err)
		reqCtx.SetDomainConstraints(domConstr)
	}
	return reqCtx
}

func (this *WatermillCqrsBus) handleRequest(
	reqCtx context.Context, handler RequestHandler, reqPacket *RequestPacket[Request], topicName string,
) (reply Reply[any], err error) {
	r, err := handler.Handle(reqCtx, reqPacket)
	if err != nil {
		this.logger.Error(
			fmt.Sprintf("error occured from topic %s", topicName),
			err,
		)
		reply.Error = util.ToPtr(err.Error())
	} else {
		reply = *r
	}
	return
}

func (this *WatermillCqrsBus) sendResponse(reqPacket *RequestPacket[Request], reply *Reply[any]) {
	replyPacket := newReplyPacket(reqPacket.correlationId, reply, this.marshaler)
	err := this.publisher.Publish(reqPacket.replyTopic, replyPacket.message)
	if err != nil {
		this.logger.Error(
			fmt.Sprintf("failed to publish reply to topic %s", reqPacket.replyTopic),
			err,
		)
	}

}

func (this *WatermillCqrsBus) RequestNoReply(ctx context.Context, request Request) (err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "send request")
	}()

	packet, err := this.newRequestPacket(ctx, request)
	ft.PanicOnErr(err)
	packet.message.Metadata.Set(MetaNoReply, "true")

	err = this.publisher.Publish(packet.requestTopic, packet.message)
	ft.PanicOnErr(err)

	return nil
}

func (this *WatermillCqrsBus) Request(ctx context.Context, request Request, result any) (err error) {
	cancellableCtx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "send request of type "+request.CqrsRequestType().String()); e != nil {
			err = e
			cancelSubscription()
		}
	}()

	packet, err := this.newRequestPacket(cancellableCtx, request)
	ft.PanicOnErr(err)

	replyChan, errChan := this.subscribeReply(cancellableCtx, packet, result, cancelSubscription)
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

func (this *WatermillCqrsBus) newRequestPacket(cancellableCtx context.Context, request Request) (packet *RequestPacket[Request], err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "create request packet for "+request.CqrsRequestType().String())
	}()
	packet = newOutgoingRequestPacket(cancellableCtx, request, this.marshaler)
	packet.message.SetContext(cancellableCtx)

	return packet, nil
}

func genRequestTopic(requestType string) string {
	return "cqrs:" + requestType
}

func genReplyTopic(requestTopic string, correlationId string) string {
	return fmt.Sprintf("%s:reply:%s", requestTopic, correlationId)
}

func newOutgoingRequestPacket(cancellableCtx context.Context, request Request, marshaler cqrs.CommandEventMarshaler) *RequestPacket[Request] {
	msg, err := marshaler.Marshal(&request)
	ft.PanicOnErr(err)

	packet := &RequestPacket[Request]{
		message: msg,
	}

	newUlid, err := ulid.New()
	ft.PanicOnErr(err)

	packet.correlationId = newUlid.String()
	requestType := request.CqrsRequestType().String()
	packet.requestTopic = genRequestTopic(requestType)
	packet.replyTopic = genReplyTopic(packet.requestTopic, packet.correlationId)

	// Caller metadata is written first and reserved keys last, so a caller-supplied value can
	// never overwrite one the bus owns. WithMetadata already filters reserved names; this
	// ordering means the guarantee holds even if it stops doing so.
	for key, val := range MetadataFrom(cancellableCtx) {
		msg.Metadata.Set(key, val)
	}

	// reqCtx, isReqCtx := cancellableCtx.(corectx.Context)
	domConstAny := cancellableCtx.Value(corectx.CtxKeyDomainConstraints)
	if domConstAny != nil {
		domConst := domConstAny.(dmodel.DynamicFields)
		domConstrJson, err := json.Marshal(domConst)
		ft.PanicOnErr(err)
		msg.Metadata.Set(MetaDomainConstraints, string(domConstrJson))
	}

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
		if e := ft.RecoverPanicFailedTo(recover(), "handle request"); e != nil {
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
