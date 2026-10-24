package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Deciding what happens to a variant the configuration no longer allows: delete it when nothing
// ever used it and the caller may delete, archive it otherwise (CR01 §7.2). "Unused" needs both
// Inventory's own history and every dependant module's answer; an unanswered check is never read
// as unused.

// Why an invalid variant is archived rather than deleted.
const (
	ArchiveReasonNoDeletePermission = "no_delete_permission"
	ArchiveReasonEverUsed           = "ever_used"
	ArchiveReasonUsedByModules      = "used_by_modules"
	ArchiveReasonHasDependents      = "has_dependent_variants"
	ArchiveReasonUsageUnknown       = "usage_unknown"
)

// ReasonUsageCheckFailed refuses an apply whose delete decisions could not be made.
const ReasonUsageCheckFailed = "product_template.usage_check_failed"

// invalidDecision is the executor's verdict on one invalid variant.
type invalidDecision struct {
	delete bool
	reason string
	usedBy []string
}

// everUsedSchemas hold a product_variant_id for every stock fact Inventory records. Any row,
// archived or zero, counts: stock at zero does not prove a variant was never used.
var everUsedSchemas = []string{
	models.StockQuantSchemaName,
	models.StockMoveSchemaName,
	models.StockMoveLineSchemaName,
	models.StockReservationSchemaName,
	models.StockScrapSchemaName,
	models.WarehouseProductGuardSchemaName,
}

// usageDecider decides delete vs archive for a batch of invalid variants.
type usageDecider struct {
	dispatcher    *usagecheck.Dispatcher
	orgId         string
	deleteAllowed bool
	// children maps a base id to the ids of every variant (any state) pointing at it.
	children map[string][]string
}

// decide answers for every invalid variant in one ever-used read per stock table and one usage
// dispatch. A dispatch error refuses the whole apply as retryable.
func (this usageDecider) decide(
	ctx corectx.Context, invalid []PlannedVariant,
) (map[string]invalidDecision, ft.ClientErrors, error) {
	decisions := map[string]invalidDecision{}
	ids := plannedVariantIds(invalid)
	if len(ids) == 0 {
		return decisions, nil, nil
	}
	if !this.deleteAllowed {
		return archiveAll(ids, ArchiveReasonNoDeletePermission), nil, nil
	}
	everUsed, err := loadEverUsedVariants(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	candidates := []string{}
	for _, id := range ids {
		if everUsed[id] {
			decisions[id] = invalidDecision{reason: ArchiveReasonEverUsed}
			continue
		}
		candidates = append(candidates, id)
	}
	cErrs, err := this.applyModuleUsage(ctx, candidates, decisions)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	this.keepReferencedBases(decisions)
	return decisions, nil, nil
}

func (this usageDecider) applyModuleUsage(
	ctx corectx.Context, candidates []string, decisions map[string]invalidDecision,
) (ft.ClientErrors, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	if this.dispatcher == nil {
		for _, id := range candidates {
			decisions[id] = invalidDecision{reason: ArchiveReasonUsageUnknown}
		}
		return nil, nil
	}
	refs := make([]usagecheck.ResourceRef, 0, len(candidates))
	for _, id := range candidates {
		refs = append(refs, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceProductVariant, Identifier: usagecheck.NewIdentifier(id, this.orgId),
		})
	}
	outcome, err := this.dispatcher.CheckUsage(withoutTransaction(ctx), modconstants.InventoryModuleName, refs)
	if err != nil {
		return nil, errors.Wrap(err, "usage check of invalid variants")
	}
	if failed := failedModules(outcome); len(failed) > 0 {
		return ft.ClientErrors{configViolation("attribute_lines", ReasonUsageCheckFailed,
			"could not confirm whether the affected variants are used ({{modules}}); try again",
			map[string]any{"modules": failed})}, nil
	}
	usedBy := usedByModules(outcome)
	for _, id := range candidates {
		if modules := usedBy[id]; len(modules) > 0 {
			decisions[id] = invalidDecision{reason: ArchiveReasonUsedByModules, usedBy: modules}
			continue
		}
		decisions[id] = invalidDecision{delete: true}
	}
	return nil, nil
}

// keepReferencedBases archives a base that would otherwise be deleted while any variant not being
// deleted still points at it: base_variant_id is RESTRICT, and history keeps its base.
func (this usageDecider) keepReferencedBases(decisions map[string]invalidDecision) {
	for changed := true; changed; {
		changed = false
		for baseId, childIds := range this.children {
			decision, found := decisions[baseId]
			if !found || !decision.delete {
				continue
			}
			for _, childId := range childIds {
				if child, ok := decisions[childId]; !ok || !child.delete {
					decisions[baseId] = invalidDecision{reason: ArchiveReasonHasDependents}
					changed = true
					break
				}
			}
		}
	}
}

// withoutTransaction detaches the request's transaction: dependants answer on their own
// connections, and a transaction is not safe to share across the dispatcher's goroutines.
func withoutTransaction(ctx corectx.Context) corectx.Context {
	detached := corectx.CloneRequestContext(ctx)
	detached.SetDbTranx(nil)
	return detached
}

func failedModules(outcome *usagecheck.CheckOutcome) []string {
	failed := []string{}
	for _, answer := range outcome.Answers {
		if answer.Err != nil {
			failed = append(failed, answer.Module)
		}
	}
	sort.Strings(failed)
	return failed
}

func usedByModules(outcome *usagecheck.CheckOutcome) map[string][]string {
	usedBy := map[string][]string{}
	for _, answer := range outcome.Answers {
		for _, item := range answer.UsedBy {
			if item.IsUsed {
				id := item.Identifier[usagecheck.IdentifierKeyId]
				usedBy[id] = append(usedBy[id], answer.Module)
			}
		}
	}
	return usedBy
}

// loadEverUsedVariants returns the ids any Inventory stock table names, one existence query per
// table for the whole batch.
func loadEverUsedVariants(ctx corectx.Context, ids []string) (map[string]bool, error) {
	used := map[string]bool{}
	keys := make([]dmodel.DynamicFields, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, dmodel.DynamicFields{"product_variant_id": id})
	}
	for _, schemaName := range everUsedSchemas {
		repo, err := repoFor(schemaName)
		if err != nil {
			return nil, err
		}
		found, err := repo.Exists(ctx, keys)
		if err != nil {
			return nil, errors.Wrap(err, "ever-used check on "+schemaName)
		}
		if found.ClientErrors.Count() > 0 {
			return nil, errors.Wrap(found.ClientErrors.ToError(), "ever-used check on "+schemaName)
		}
		for _, key := range found.Data.Existing {
			used[readStringParam(key, "product_variant_id")] = true
		}
	}
	return used, nil
}

func archiveAll(ids []string, reason string) map[string]invalidDecision {
	decisions := map[string]invalidDecision{}
	for _, id := range ids {
		decisions[id] = invalidDecision{reason: reason}
	}
	return decisions
}

func plannedVariantIds(items []PlannedVariant) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.VariantId != "" {
			ids = append(ids, item.VariantId)
		}
	}
	return ids
}
