package external

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ActionOutcome is what one attempt produced, expressed in the terms the retry policy needs
// rather than in the terms of whichever transport produced it.
//
// Reducing an HTTP status or a command error to Succeeded plus Retryable at this boundary is what
// keeps the retry policy free of transport knowledge: it decides when to retry, never how to tell
// whether something failed.
type ActionOutcome struct {
	Succeeded bool

	// Retryable is meaningless when Succeeded. It reports whether repeating the action could
	// plausibly give a different answer.
	Retryable bool

	// ErrorCode is machine-readable and stable: "HTTP_503", "TIMEOUT", "COMMAND_ERROR".
	ErrorCode string

	// ErrorMessage is for a human reading the history. It must be sanitized: no credentials, and
	// no full URLs with query strings, since an action config can carry a token in either.
	ErrorMessage string

	// HttpStatusCode is set only by the REST executor.
	HttpStatusCode *int32
}

// ActionInput is everything an executor needs to run one attempt.
type ActionInput struct {
	Config map[string]any

	// ExecutionKey identifies the occurrence rather than the attempt, so every attempt of the same
	// execution presents the same idempotency key. That is the point: a retry must be recognizable
	// to the receiver as the same work, not as new work.
	ExecutionKey string

	JobId         string
	ExecutionId   string
	AttemptNumber int

	// Timeout bounds this attempt. It comes from application configuration rather than the job,
	// and is applied by the executor to whatever transport it uses.
	Timeout time.Duration
}

// ActionExecutor runs one attempt of one action type.
type ActionExecutor interface {
	// ActionType is the discriminator stored on the job.
	ActionType() string

	// Validate checks an action config when the job is created, returning field-scoped client
	// errors rather than an error so the REST layer can answer 400 with a reason a caller can act
	// on. Validating here rather than at first run means a misconfigured job is refused at
	// registration instead of failing silently at 3am.
	Validate(ctx corectx.Context, config map[string]any) *ft.ClientErrors

	// Execute runs the action. It returns an outcome rather than an error because every failure
	// mode here is a business outcome to be recorded on the attempt, not an exception: a 500 from
	// a remote service is data, not a bug in the scheduler.
	Execute(ctx context.Context, in ActionInput) ActionOutcome
}
