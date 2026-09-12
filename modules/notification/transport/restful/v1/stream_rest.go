package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// StreamRest serves the NDJSON notification stream.
//
// It is registered as a hand-written route rather than through the resource engine, because the
// engine answers one JSON document per request and this answers a response that never ends. That
// is the documented exception for an unusual HTTP surface.
type StreamRest struct {
	inbox    it.InboxApplicationService
	registry it.StreamRegistry
	cfg      config.ConfigService
	logger   logging.LoggerService
}

func NewStreamRest(
	inbox it.InboxApplicationService,
	registry it.StreamRegistry,
	cfg config.ConfigService,
	logger logging.LoggerService,
) *StreamRest {
	return &StreamRest{inbox: inbox, registry: registry, cfg: cfg, logger: logger}
}

// Stream holds one HTTP response open and writes notifications to it as they arrive.
//
// The order is load-bearing: replay whatever the client missed FIRST, then go live (BR-FS 10,
// BR-FS 11). Registering with the stream registry before the replay read is equally deliberate —
// a notification committed during the replay would otherwise fall between the two and be lost,
// which is the one gap this design has to close.
func (this *StreamRest) Stream(echoCtx *echo.Context, _ map[string]any) error {
	ctx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}

	scope, err := this.inbox.AuthorizeStream(ctx)
	if err != nil {
		return (*echoCtx).JSON(http.StatusForbidden, map[string]any{"error": err.Error()})
	}

	buffer := int(this.cfg.GetInt32(modconstants.StreamMaxBufferedEvents, "100"))
	events, unregister := this.registry.Register(scope.OrgId, scope.UserId, buffer)
	defer unregister()

	this.writeStreamHeaders(echoCtx)

	afterSeq := parseAfterSeq(*echoCtx)
	lastSeq, err := this.replay(ctx, echoCtx, afterSeq)
	if err != nil {
		return err
	}

	return this.pump(ctx, echoCtx, events, lastSeq)
}

// writeStreamHeaders declares a response that must not be buffered or cached anywhere between
// here and the browser, and that has no length because it has no end (BR-FS 3).
func (this *StreamRest) writeStreamHeaders(echoCtx *echo.Context) {
	response := (*echoCtx).Response()
	header := response.Header()
	header.Set(echo.HeaderContentType, "application/x-ndjson")
	header.Set("Cache-Control", "no-cache, no-transform")
	// Belt and braces for a proxy that buffers by default. Traefik does not, but the contract is
	// that this endpoint works behind any gateway without one being specially configured.
	header.Set("X-Accel-Buffering", "no")

	// No Content-Length: the response has no end (BR-FS 3). Writing the header now rather than on
	// the first event is what gets the 200 to the client immediately, so it can start reading.
	response.WriteHeader(http.StatusOK)
	flush(response)
}

// flush pushes what has been written out to the client.
//
// http.NewResponseController rather than a direct http.Flusher assertion, because the writer is
// wrapped by middleware and the controller unwraps to find the flusher. A writer that cannot flush
// is a deployment that cannot stream, and silently buffering would look like notifications simply
// never arriving.
func flush(response http.ResponseWriter) {
	_ = http.NewResponseController(response).Flush()
}

// replay writes everything newer than the client's cursor, and returns the highest sequence it
// sent so the live phase can skip what the client has already seen.
func (this *StreamRest) replay(
	ctx corectx.Context, echoCtx *echo.Context, afterSeq *int64,
) (int64, error) {
	pageSize := int(this.cfg.GetInt32(modconstants.StreamReplayPageSize, "200"))

	result, err := this.inbox.Replay(ctx, it.StreamQuery{AfterSeq: afterSeq}, pageSize)
	if err != nil {
		return 0, err
	}
	if result.ClientErrors.Count() > 0 {
		return 0, result.ClientErrors.ToError()
	}

	var lastSeq int64
	if afterSeq != nil {
		lastSeq = *afterSeq
	}
	for _, item := range result.Data.Items {
		seq := item.StreamSeq
		event := it.StreamEvent{
			Seq:  &seq,
			Type: it.StreamEventNotificationCreated,
			Data: toInboxItemDto(item),
		}
		if err := this.writeEvent(echoCtx, event); err != nil {
			return lastSeq, nil
		}
		lastSeq = seq
	}
	return lastSeq, nil
}

// pump is the live phase: events until the client goes away, with a heartbeat whenever it falls
// quiet.
//
// The heartbeat is not decoration. It keeps traffic on an idle connection so that a proxy does not
// close it as dead, and it is how this end notices a client that vanished without a FIN (BR-FS 9).
func (this *StreamRest) pump(
	ctx corectx.Context, echoCtx *echo.Context, events <-chan it.StreamEvent, lastSeq int64,
) error {
	interval := time.Duration(this.cfg.GetInt32(
		modconstants.StreamHeartbeatIntervalSecs, "20")) * time.Second
	heartbeat := time.NewTicker(interval)
	defer heartbeat.Stop()

	request := (*echoCtx).Request()

	for {
		select {
		case <-request.Context().Done():
			// The client disconnected. That is not a delivery failure and must leave the
			// notification and its read state exactly as they were (BR-FS 21).
			return nil

		case event, open := <-events:
			if !open {
				return nil
			}
			// A replayed notification may also arrive live. Skipping what the client has already
			// been sent keeps the duplicate out rather than relying on it to dedupe (BR-FS 12).
			if event.Seq != nil && *event.Seq <= lastSeq {
				continue
			}
			if err := this.writeEvent(echoCtx, event); err != nil {
				return nil
			}
			if event.Seq != nil {
				lastSeq = *event.Seq
			}

		case <-heartbeat.C:
			if err := this.writeEvent(echoCtx, it.StreamEvent{Type: it.StreamEventHeartbeat}); err != nil {
				return nil
			}
		}
	}
}

// writeEvent writes one NDJSON line and flushes it.
//
// Flushing every line is the whole point of the transport: an event held in a buffer waiting for
// company is not realtime (BR-FS 8). A write error means the client is gone, and the caller ends
// the stream rather than retrying.
func (this *StreamRest) writeEvent(echoCtx *echo.Context, event it.StreamEvent) error {
	line, err := json.Marshal(event)
	if err != nil {
		// A notification whose payload will not marshal must not take the stream down with it:
		// every other notification on this connection is still deliverable.
		this.logger.Error("notification stream event is not serializable", err)
		return nil
	}

	response := (*echoCtx).Response()
	if _, err := response.Write(append(line, '\n')); err != nil {
		return err
	}
	flush(response)
	return nil
}

// parseAfterSeq reads the client's cursor. An unreadable value is treated as absent, which starts
// the client from the live edge rather than replaying its whole history by accident.
func parseAfterSeq(echoCtx echo.Context) *int64 {
	raw := strings.TrimSpace(echoCtx.QueryParam("after_seq"))
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &value
}
