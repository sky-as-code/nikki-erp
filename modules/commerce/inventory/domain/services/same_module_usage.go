package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// SameModulePart is another part of Commerce that references Inventory resources: Purchase and
// Sales name variants and templates on their documents and pricing rules. Being parts of the same
// module, they are no dependants and answer in process, from the registry they build, keyed by
// (CommerceModuleName, resource).
type SameModulePart struct {
	Name     string
	Registry *usagecheck.CheckerRegistry
}

// sameModuleParts is installed once by Commerce's Init, before any request. Like the outbox
// drain, it is a package hook because the parts that answer are built after Inventory.
var sameModuleParts []SameModulePart

// SetSameModuleParts installs the parts Inventory asks before deleting a resource they may name.
func SetSameModuleParts(parts ...SameModulePart) {
	sameModuleParts = parts
}

// checkUsage asks Commerce's dependants over the bus and its other parts in process, and returns
// one outcome holding both. A part's answer is named after the part, so a refusal tells the user
// where to release the reference exactly as a dependant's does.
func checkUsage(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher, refs []usagecheck.ResourceRef,
) (*usagecheck.CheckOutcome, error) {
	outcome, err := dispatcher.CheckUsage(ctx, commconstants.CommerceModuleName, refs)
	if err != nil {
		return nil, err
	}
	for _, part := range sameModuleParts {
		outcome.Answers = append(outcome.Answers, answerInProcess(ctx, part, refs))
	}
	return outcome, nil
}

// answerInProcess asks one part about every resource. A resource the part registered no checker
// for is an error, never "unused": the part would otherwise silently stop guarding it.
func answerInProcess(
	ctx corectx.Context, part SameModulePart, refs []usagecheck.ResourceRef,
) (answer usagecheck.ModuleAnswer) {
	answer.Module = part.Name
	for _, ref := range refs {
		checker, found := part.Registry.Checker(commconstants.CommerceModuleName, ref.ResourceName)
		if !found {
			answer.Err = errors.Errorf("commerce part %s has no checker for %s", part.Name, ref.ResourceName)
			return answer
		}
		used, usedBy, err := checker.IsUsed(ctx, ref)
		if err != nil {
			answer.Err = errors.Wrapf(err, "usage check on commerce part %s", part.Name)
			return answer
		}
		if used {
			answer.UsedBy = append(answer.UsedBy, usagecheck.ResourceUsageItem{
				ResourceName: ref.ResourceName, Identifier: ref.Identifier, IsUsed: true, UsedBy: usedBy,
			})
		}
	}
	answer.Used = len(answer.UsedBy) > 0
	return answer
}

// assertDeletable is usagecheck.AssertDeletable over checkUsage: it refuses the delete when a
// dependant or another part of Commerce still names the resource.
func assertDeletable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	idField string, ref usagecheck.ResourceRef, vErrs *ft.ClientErrors,
) error {
	outcome, err := checkUsage(ctx, dispatcher, []usagecheck.ResourceRef{ref})
	if err != nil {
		return err
	}
	usagecheck.AppendBlockingViolation(outcome, idField, ref, vErrs)
	return nil
}
