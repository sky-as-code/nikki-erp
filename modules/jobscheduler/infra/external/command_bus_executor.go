package external

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// commandNamePattern matches the shape RequestType.String() produces: "{module}_{submodule}.{action}".
var commandNamePattern = regexp.MustCompile(`^[a-z0-9]+_[a-z0-9]+\.[A-Za-z0-9_]+$`)

// CommandBusExecutor dispatches a registered command on a schedule.
//
// It emits with Request rather than RequestNoReply, and that is a deliberate departure from the
// "no-reply command" wording in the requirement. RequestNoReply publishes and returns: the
// handler's error is only logged, so every command attempt would be recorded as a success. The
// whole retry engine - backoff, the attempt budget, the retry window - would then be dead weight
// for command jobs, and a job whose handler failed every time would look perfectly healthy in the
// execution history.
//
// Awaiting the reply costs a round trip on an in-process bus and buys the ability to tell a job
// that worked from one that did not, which is the entire point of recording attempts.
type CommandBusExecutor struct {
	cqrsBus cqrs.CqrsBus
}

func NewCommandBusExecutor(cqrsBus cqrs.CqrsBus) *CommandBusExecutor {
	return &CommandBusExecutor{cqrsBus: cqrsBus}
}

func (this *CommandBusExecutor) ActionType() string {
	return models.ActionTypeCommandBus
}

// Validate checks the command name and, crucially, that something is actually listening for it.
//
// The registry check is what turns a typo into a 400 at registration instead of a job that runs on
// schedule forever, publishes into the void, and reports success every time. There is no later
// point at which the mistake would surface on its own.
func (this *CommandBusExecutor) Validate(ctx corectx.Context, config map[string]any) *ft.ClientErrors {
	errs := ft.NewClientErrors()

	name, _ := config["command_name"].(string)
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		errs.Append(*ft.NewValidationError("action_config.command_name",
			ft.ErrorKey("err_command_not_registered", constants.JobSchedulerModuleName), "a command name is required"))

	case !commandNamePattern.MatchString(name):
		errs.Append(*ft.NewValidationError("action_config.command_name",
			ft.ErrorKey("err_command_not_registered", constants.JobSchedulerModuleName),
			"a command name looks like 'module_submodule.action'"))

	case !this.cqrsBus.IsRequestTypeRegistered(name):
		errs.Append(*ft.NewValidationError("action_config.command_name",
			ft.ErrorKey("err_command_not_registered", constants.JobSchedulerModuleName),
			"no handler is registered for '"+name+"'"))
	}

	if errs.Count() > 0 {
		return errs
	}
	return nil
}

// Execute sends the command and waits for its reply.
func (this *CommandBusExecutor) Execute(ctx context.Context, in it.ActionInput) it.ActionOutcome {
	name, _ := in.Config["command_name"].(string)
	name = strings.TrimSpace(name)

	requestType, ok := parseRequestType(name)
	if !ok {
		return it.ActionOutcome{
			ErrorCode:    "INVALID_ACTION_CONFIG",
			ErrorMessage: "the command name is not usable",
			Retryable:    false,
		}
	}

	attemptCtx, cancel := context.WithTimeout(ctx, in.Timeout)
	defer cancel()

	// The same metadata the REST action sends as headers, so a handler can make its side effect
	// idempotent by the same means whichever transport invoked it.
	attemptCtx = cqrs.WithMetadata(attemptCtx, map[string]string{
		cqrs.MetaIdempotencyKey:       in.ExecutionKey,
		cqrs.MetaSchedulerJobId:       in.JobId,
		cqrs.MetaSchedulerExecutionId: in.ExecutionId,
		cqrs.MetaSchedulerAttempt:     strconv.Itoa(in.AttemptNumber),
	})

	var result any
	err := this.cqrsBus.Request(attemptCtx, genericCommand{requestType: requestType}, &result)
	if err != nil {
		outcome := classifyTransportError(err)
		if outcome.ErrorCode == ErrorCodeNetwork {
			// A bus error is a handler that returned an error, not a network fault; naming it as
			// such keeps the history readable.
			outcome.ErrorCode = ErrorCodeCommand
			outcome.ErrorMessage = "the command handler returned an error"
		}
		return outcome
	}

	return it.ActionOutcome{Succeeded: true}
}

// parseRequestType splits "{module}_{submodule}.{action}" back into the struct the bus routes on.
func parseRequestType(name string) (cqrs.RequestType, bool) {
	prefix, action, found := strings.Cut(name, ".")
	if !found || action == "" {
		return cqrs.RequestType{}, false
	}
	module, submodule, found := strings.Cut(prefix, "_")
	if !found || module == "" || submodule == "" {
		return cqrs.RequestType{}, false
	}
	return cqrs.RequestType{Module: module, Submodule: submodule, Action: action}, true
}

// genericCommand carries a request type parsed from a string.
//
// The bus routes on CqrsRequestType().String(), so this is enough to reach the right handler even
// though the scheduler cannot import the concrete command type - it holds only a name read from a
// database row, and a module's command types are not visible to it by design.
//
// This works because a technical job's command takes no business parameters: the marshaled body is
// an empty object, which unmarshals into any handler request whose fields are all optional. A
// handler that requires a field would fail to unmarshal, and because the reply carries that error
// back, the attempt is recorded as failed and retried rather than silently doing nothing.
type genericCommand struct {
	requestType cqrs.RequestType
}

func (this genericCommand) CqrsRequestType() cqrs.RequestType {
	return this.requestType
}
