package composable

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The repository-level trail covers what a custom domain method writes straight through
// Repository().Insert/Update/DeleteOne: an order's lines, a stock move, a status change. Each
// write joins or opens a transaction with its entry, like the domain-level trail. An update reads
// the row first to diff it, the price of covering writes the domain service never sees.

func (this *auditTrail) insert(
	ctx corectx.Context, data dmodel.DynamicFields,
	run func(ctx corectx.Context, data dmodel.DynamicFields) (*dyn.OpResult[int], error),
) (*dyn.OpResult[int], error) {
	var result *dyn.OpResult[int]
	err := InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		var err error
		result, err = run(tranxCtx, data)
		if err != nil || result == nil || result.ClientErrors.Count() > 0 {
			return err
		}
		return this.record(tranxCtx, audit.ActionCreate, nil, data)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// updateRow records nothing for a row it cannot address by id, such as a junction row keyed by
// its pair: there is no record for a history to belong to.
func (this *auditTrail) updateRow(
	ctx corectx.Context, data dmodel.DynamicFields,
	run func(ctx corectx.Context, data dmodel.DynamicFields) (*MutateResult, error),
) (*MutateResult, error) {
	var result *MutateResult
	err := InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		before, err := this.storedRow(tranxCtx, data)
		if err != nil {
			return err
		}
		result, err = run(tranxCtx, data)
		if err != nil || !removedOrChanged(result) || before == nil {
			return err
		}
		return this.record(tranxCtx, updateActionOf(tranxCtx), before, data)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
