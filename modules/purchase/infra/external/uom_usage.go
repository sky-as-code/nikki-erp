package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Purchase's answer to "is this unit of measure in use" (BR-UOM-ESS-020).
//
// Essential refuses to change a unit's factor, type or category while transactions reference it,
// because doing so would reinterpret quantities already recorded. Until now nothing consumed UoM,
// so its probe was a stub returning false and the rule was enforced against an assumption.
// Purchase is the first real consumer, and this is what makes the rule true.

// RegisterUomUsageProbe tells Essential that Purchase holds unit references.
func RegisterUomUsageProbe() {
	itUom.RegisterUomUsageProbe(&purchaseUomProbe{})
}

type purchaseUomProbe struct{}

var _ itUom.UomUsageProbe = (*purchaseUomProbe)(nil)

func (*purchaseUomProbe) ModuleName() string {
	return "purchase"
}

// IsUomInUse reports whether any purchase order line or agreement line names this unit.
//
// Both are checked, and an agreement line counts: an agreement commits to a quantity at a price in
// a stated unit, so reinterpreting that unit would change what the business agreed to buy just as
// surely as changing a confirmed order would.
//
// A read that fails returns the error rather than false. Essential treats a failed probe as "in
// use" and refuses the edit, which is the safe direction — refusing an edit that might have been
// fine is recoverable, silently changing what a historical document means is not.
func (this *purchaseUomProbe) IsUomInUse(ctx corectx.Context, uomId string) (bool, error) {
	if uomId == "" {
		return false, nil
	}

	for _, target := range []struct {
		schemaName string
		field      string
	}{
		{models.PurchaseOrderLineSchemaName, models.PurchaseOrderLineFieldUomId},
		{models.AgreementLineSchemaName, models.AgreementLineFieldUomId},
	} {
		inUse, err := anyReferencing(ctx, target.schemaName, target.field, uomId)
		if err != nil {
			return false, err
		}
		if inUse {
			return true, nil
		}
	}
	return false, nil
}

// anyReferencing reports whether one resource has at least one row whose field holds the value.
//
// Size 1: the question is "any", so reading a second row would cost a page for an answer already
// settled by the first.
func anyReferencing(
	ctx corectx.Context, schemaName, field, value string,
) (bool, error) {
	engine, ok := engineFor(schemaName)
	if !ok {
		// The engine is missing only if this module failed to initialise, in which case it holds
		// no data either. Reporting "not in use" is accurate rather than optimistic.
		return false, nil
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
		// Archived lines still count: an archived agreement's lines are history, and history is
		// exactly what must not be reinterpreted.
		IncludeArchived: nil,
	})
	if err != nil {
		return false, errors.Wrapf(err, "checking whether %s references %s", schemaName, field)
	}
	return found != nil && found.HasData && len(found.Data.Items) > 0, nil
}
