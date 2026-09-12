package usagecheck

import (
	"strings"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules"
)

// ErrorKeyResourceInUse is the refusal a blocked delete reports. One key for every resource, so
// a client can recognize the condition without matching on message text.
const ErrorKeyResourceInUse = "resource_in_use"

// AssertDeletable runs the usage check and appends a violation when any dependant blocks.
//
// It returns an error only when the check could not be run at all. A refusal is a client error,
// not a failure: it is the expected answer to "delete a unit something still measures".
//
// The message names the modules that blocked, because that is the actionable part - the user
// must go to those modules to release the reference - but never which records they hold, which
// is the dependant's data and may be outside this user's visibility.
func AssertDeletable(
	ctx corectx.Context,
	dispatcher *Dispatcher,
	sourceModule string,
	idField string,
	ref ResourceRef,
	vErrs *ft.ClientErrors,
) error {
	outcome, err := dispatcher.CheckUsage(ctx, sourceModule, []ResourceRef{ref})
	if err != nil {
		return err
	}
	if outcome.MayDelete() {
		return nil
	}

	blocking := outcome.BlockingModules()
	vErrs.Append(*ft.NewBusinessViolation(
		idField,
		ErrorKeyResourceInUse,
		"this "+ref.ResourceName+" is still used by "+strings.Join(blocking, ", ")+
			"; remove or replace those references first",
		map[string]any{
			"resource_name": ref.ResourceName,
			"used_by":       blocking,
		},
	))
	return nil
}

// AssertDependantsSubscribed fails startup when a loaded dependant never subscribed a usage
// handler.
//
// Without it the gap is silent and fails open in the worst way: the owning module dispatches a
// check, the bus finds no handler, the request times out, and every delete of that resource is
// refused with an infrastructure error - or, if the dispatch were ever made lenient, permitted
// with a reference still live. Boot is the right place to notice, since subscriptions happen in
// module Init and are complete by the time OnAppStarted runs.
func AssertDependantsSubscribed(
	bus cqrs.CqrsBus, dependants *modules.ModuleDependantRegistry, ownerModule string,
) error {
	missing := make([]string, 0)
	for _, dependant := range dependants.GetModuleDependants(ownerModule) {
		if !bus.IsRequestTypeRegistered(RequestTypeFor(dependant).String()) {
			missing = append(missing, dependant)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf(
			"module %s lists dependant(s) %s, which did not subscribe a usage check handler; "+
				"either register a checker in those modules or remove them from Dependants()",
			ownerModule, strings.Join(missing, ", "))
	}
	return nil
}
