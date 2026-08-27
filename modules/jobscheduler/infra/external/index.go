package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/services"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// ActionExecutors is the set of executors keyed by the action type each handles.
//
// It is a map rather than a slice so that dispatch is a lookup rather than a scan, and so that two
// executors claiming the same action type is impossible by construction rather than a first-match
// surprise.
type ActionExecutors map[string]it.ActionExecutor

// InitExternalServices registers the executors and the map that dispatches between them.
//
// This is the only place the scheduler touches other subsystems: the command bus and the HTTP
// client. Keeping the seam here is what lets the domain services stay free of transport knowledge.
func InitExternalServices() error {
	return stdErr.Join(
		deps.Register(NewCommandBusExecutor),
		deps.Register(NewRestApiExecutor),
		deps.Register(newActionExecutors),
		deps.Register(NewActionDispatcher),
		// The dispatcher is what the domain service validates action configs through. It is
		// registered under the domain's interface too, so that layer never imports infra.
		deps.Register(func(d *ActionDispatcher) services.ActionConfigValidator { return d }),
	)
}

func newActionExecutors(
	commandBus *CommandBusExecutor, restApi *RestApiExecutor,
) ActionExecutors {
	return ActionExecutors{
		commandBus.ActionType(): commandBus,
		restApi.ActionType():    restApi,
	}
}

// Verify at compile time that both executors satisfy the contract, rather than discovering a
// mismatch when the container is built at start-up.
var (
	_ it.ActionExecutor = (*CommandBusExecutor)(nil)
	_ it.ActionExecutor = (*RestApiExecutor)(nil)
)
