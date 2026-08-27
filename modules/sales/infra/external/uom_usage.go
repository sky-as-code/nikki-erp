package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Sales' answer to "is this unit of measure in use" (BR-UOM-ESS-020).
//
// Essential refuses to change a unit's factor, type or category while transactions reference it,
// because doing so would reinterpret quantities already recorded. Every sales order line carries a
// uom_id, so without this probe Essential would allow editing a unit that sales history depends on
// — and a receipt issued last year would silently come to mean a different amount of goods.

// RegisterUomUsageProbe tells Essential that Sales holds unit references.
func RegisterUomUsageProbe() {
	itUom.RegisterUomUsageProbe(&salesUomProbe{})
}

type salesUomProbe struct{}

var _ itUom.UomUsageProbe = (*salesUomProbe)(nil)

func (*salesUomProbe) ModuleName() string {
	return "sales"
}

// IsUomInUse reports whether any sales order line names this unit.
//
// Only the line table is checked today because it is the only Sales schema carrying a uom_id. The
// list is a slice rather than a single name so that the combo components of SALES-008 and the return
// lines of SALES-032 — both of which will carry one — are added by appending an entry rather than by
// reshaping the function.
//
// A read that fails returns the error rather than false. Essential treats a failed probe as "in use"
// and refuses the edit, which is the safe direction: refusing an edit that might have been fine is
// recoverable, silently changing what a historical document means is not.
func (this *salesUomProbe) IsUomInUse(ctx corectx.Context, uomId string) (bool, error) {
	if uomId == "" {
		return false, nil
	}

	for _, target := range []struct {
		schemaName string
		field      string
	}{
		{models.SalesOrderLineSchemaName, models.SalesOrderLineFieldUomId},
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
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		// The engine is missing only if this module failed to initialise, in which case it holds no
		// data either. Reporting "not in use" is accurate rather than optimistic.
		return false, nil
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
		// Archived lines still count: an archived order's lines are history, and history is exactly
		// what must not be reinterpreted.
		IncludeArchived: nil,
	})
	if err != nil {
		return false, errors.Wrapf(err, "checking whether %s references %s", schemaName, field)
	}
	return found != nil && found.HasData && len(found.Data.Items) > 0, nil
}
