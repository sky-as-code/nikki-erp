package external

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// ActionDispatcher routes an action to the executor that handles its type.
//
// It is the one place that knows the executor set, so that neither the domain service (which
// validates a config) nor the engine (which runs one) has to. Both ask the same map, which is
// what stops a job passing validation under one executor and being run by another.
type ActionDispatcher struct {
	executors ActionExecutors
}

func NewActionDispatcher(executors ActionExecutors) *ActionDispatcher {
	return &ActionDispatcher{executors: executors}
}

// ValidateActionConfig checks a config against the executor for its type.
//
// An unrecognized action type is an error rather than a pass. Accepting one would create a job
// that is registered, scheduled, and unrunnable - and it would fail at every occurrence rather
// than once at registration, which is the only moment somebody is watching.
func (this *ActionDispatcher) ValidateActionConfig(
	actionType string, config map[string]any,
) *ft.ClientErrors {
	executor, found := this.executors[actionType]
	if !found {
		errs := ft.NewClientErrors()
		errs.Append(*ft.NewValidationError(
			models.JobFieldActionType,
			ft.ErrorKey("err_action_type_unknown", constants.JobSchedulerModuleName),
			"no executor handles this action type",
		))
		return errs
	}
	// The config is passed even when nil: an executor's own validation is what reports a missing
	// required key, and short-circuiting here would replace its specific message with silence.
	return executor.Validate(nil, config)
}

// ExecutorFor answers the executor for an action type, or nil.
//
// The engine uses this at run time. A nil answer there means a job was registered when an
// executor existed and is being run after it was removed, which is a deployment change rather
// than a caller's mistake, and so is reported as a failed attempt rather than a panic.
func (this *ActionDispatcher) ExecutorFor(actionType string) it.ActionExecutor {
	return this.executors[actionType]
}

// Verify the dispatcher satisfies what the domain service asks of it. The domain package cannot
// import this one, so without this the mismatch would only appear when the container is built.
var _ interface {
	ValidateActionConfig(actionType string, config map[string]any) *ft.ClientErrors
} = (*ActionDispatcher)(nil)
