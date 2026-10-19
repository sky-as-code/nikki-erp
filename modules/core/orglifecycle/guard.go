package orglifecycle

import (
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// AssertSubscribed fails startup when a loaded module never subscribed its lifecycle handlers.
// Without it the gap shows up only when an org is created: the command times out and the org
// stays in provisioning. Call it from OnAppStarted, after every module's Init.
func AssertSubscribed(bus cqrs.CqrsBus, moduleNames []string) error {
	missing := make([]string, 0)
	for _, name := range moduleNames {
		for _, action := range []string{ActionInitOrg, ActionCleanUpOrg} {
			if !bus.IsRequestTypeRegistered(RequestTypeFor(name, action).String()) {
				missing = append(missing, name+"."+action)
			}
		}
	}

	if len(missing) > 0 {
		return errors.Errorf(
			"module(s) did not subscribe org lifecycle handlers: %s; "+
				"call orglifecycle.RegisterHandlers in each module's transport/cqrs Init",
			strings.Join(missing, ", "))
	}
	return nil
}
