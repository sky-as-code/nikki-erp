package external

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// Headers the scheduler sends on every REST action, so a receiver can make its side effect
// idempotent and can trace a call back to the occurrence that made it.
const (
	HeaderIdempotencyKey = "Idempotency-Key"
	HeaderJobId          = "X-Scheduler-Job-ID"
	HeaderExecutionId    = "X-Scheduler-Execution-ID"
	HeaderAttempt        = "X-Scheduler-Attempt"
)

// maxResponseBytes caps what is read from a response. The scheduler needs the status, not the
// body, and an action pointed at something that streams would otherwise hold a worker and its
// memory for as long as the peer cared to keep writing.
const maxResponseBytes = 64 * 1024

// supportedMethods is the exact set the requirement allows. It is a fixed list rather than a
// permissive check because an unexpected verb reaching an internal endpoint is a way to cause an
// effect nobody intended.
var supportedMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
	http.MethodHead:   true,
}

// RestApiExecutor calls an HTTP endpoint on a schedule.
//
// It uses the shared *client.HttpClient directly rather than httpclient.NewHttpCaller, because the
// caller requires a fixed base URL and panics on a malformed one, while a rest_api action carries
// a full absolute URL that arrives from the API. Constructing a caller per action would turn a bad
// URL into a panic on a worker goroutine.
type RestApiExecutor struct {
	httpClient *client.HttpClient
}

func NewRestApiExecutor(httpClient *client.HttpClient) *RestApiExecutor {
	return &RestApiExecutor{httpClient: httpClient}
}

func (this *RestApiExecutor) ActionType() string {
	return models.ActionTypeRestApi
}

// Validate checks the action config at job-creation time.
func (this *RestApiExecutor) Validate(ctx corectx.Context, config map[string]any) *ft.ClientErrors {
	errs := ft.NewClientErrors()

	method, _ := config["method"].(string)
	if method == "" {
		errs.Append(*ft.NewValidationError("action_config.method",
			ft.ErrorKey("err_http_method_unsupported", constants.JobSchedulerModuleName), "an HTTP method is required"))
	} else if !supportedMethods[method] {
		// Compared exactly rather than case-insensitively: accepting "post" and sending "POST"
		// would mean the job does not say what it does.
		errs.Append(*ft.NewValidationError("action_config.method",
			ft.ErrorKey("err_http_method_unsupported", constants.JobSchedulerModuleName),
			"method must be one of GET, POST, PUT, PATCH, DELETE, HEAD in upper case"))
	}

	rawUrl, _ := config["url"].(string)
	if err := validateActionUrl(rawUrl); err != "" {
		errs.Append(*ft.NewValidationError("action_config.url", ft.ErrorKey("err_url_invalid", constants.JobSchedulerModuleName), err))
	}

	if errs.Count() > 0 {
		return errs
	}
	return nil
}

// validateActionUrl returns a reason the URL is unusable, or an empty string.
//
// Restricting the scheme to http and https is a security decision rather than a tidiness one: the
// URL comes from an API caller, and file:// or gopher:// would turn an authenticated job
// registration into a way to read local files or reach services never meant to be reachable.
func validateActionUrl(rawUrl string) string {
	if strings.TrimSpace(rawUrl) == "" {
		return "a URL is required"
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "the URL could not be parsed"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "the URL must use http or https"
	}
	if parsed.Host == "" {
		return "the URL must be absolute and include a host"
	}
	return ""
}

// Execute performs one HTTP attempt.
func (this *RestApiExecutor) Execute(ctx context.Context, in it.ActionInput) it.ActionOutcome {
	method, _ := in.Config["method"].(string)
	rawUrl, _ := in.Config["url"].(string)

	// The config was validated at creation, but a row can outlive a validation change, so a bad
	// value here fails the attempt rather than panicking a worker.
	if !supportedMethods[method] || validateActionUrl(rawUrl) != "" {
		return it.ActionOutcome{
			ErrorCode:    "INVALID_ACTION_CONFIG",
			ErrorMessage: "the action config is not usable",
			Retryable:    false,
		}
	}

	// The timeout comes from configuration and is applied per attempt. The shared client's own
	// Timeout is a separate, longer ceiling covering the whole request; relying on it would let an
	// attempt outlive the lease that protects it.
	attemptCtx, cancel := context.WithTimeout(ctx, in.Timeout)
	defer cancel()

	request, err := this.buildRequest(attemptCtx, method, rawUrl, in)
	if err != nil {
		return it.ActionOutcome{
			ErrorCode:    "INVALID_ACTION_CONFIG",
			ErrorMessage: "the request could not be built",
			Retryable:    false,
		}
	}

	response, err := this.httpClient.Do(request)
	if err != nil {
		return ClassifyRestResult(0, err)
	}
	defer response.Body.Close()

	// Drain a bounded prefix so the connection can be reused, and discard it: the scheduler
	// records the status, not the payload.
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))

	return ClassifyRestResult(response.StatusCode, nil)
}

func (this *RestApiExecutor) buildRequest(
	ctx context.Context, method string, rawUrl string, in it.ActionInput,
) (*http.Request, error) {
	var body io.Reader
	if raw, ok := in.Config["body"].(string); ok && raw != "" {
		body = bytes.NewReader([]byte(raw))
	}

	request, err := http.NewRequestWithContext(ctx, method, rawUrl, body)
	if err != nil {
		return nil, err
	}

	// The action's own headers go on first so that the scheduler's cannot be overridden below.
	if configured, ok := in.Config["headers"].(map[string]any); ok {
		for name, value := range configured {
			if text, ok := value.(string); ok {
				request.Header.Set(name, text)
			}
		}
	}

	// Set last, deliberately. These identify the occurrence, and an action that could rewrite them
	// could make two different occurrences look like the same work to the receiver - defeating the
	// idempotency they exist to provide.
	request.Header.Set(HeaderIdempotencyKey, in.ExecutionKey)
	request.Header.Set(HeaderJobId, in.JobId)
	request.Header.Set(HeaderExecutionId, in.ExecutionId)
	request.Header.Set(HeaderAttempt, strconv.Itoa(in.AttemptNumber))

	if body != nil && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}

	return request, nil
}
