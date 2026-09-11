package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Sales' answer to "is this unit of measure in use". Essential refuses to change a unit's factor,
// type or category while transactions reference it, because that would reinterpret quantities
// already recorded. Without this probe a receipt issued last year could silently come to mean a
// different amount of goods.

// RegisterUomUsageProbe tells Essential that Sales holds unit references.
func RegisterUomUsageProbe() {
	itUom.RegisterUomUsageProbe(&salesUomProbe{})
}

type salesUomProbe struct{}

var _ itUom.UomUsageProbe = (*salesUomProbe)(nil)

func (*salesUomProbe) ModuleName() string {
	return "sales"
}

// IsUomInUse reports whether any sales order line names this unit. Only the line table is checked
// today because it is the only Sales schema carrying a uom_id; the slice lets later ones be added by
// appending. A failed read returns the error rather than false: essential treats a failed probe as
// "in use" and refuses the edit, which is the safe direction.
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
// Size 1: the question is "any", so a second row costs a page for an answer already settled.
func anyReferencing(
	ctx corectx.Context, schemaName, field, value string,
) (bool, error) {
	// Through the resource hub, which answers whichever generation serves the schema. Resolved at
	// probe-call time rather than at registration, so the eager registration in InitExternal stays
	// correct even though the onions are built later.
	//
	// The "not in use" answer on a failed lookup is preserved rather than fixed here: it is a
	// filed defect (ESS-2026-09-11-001 in docs/problems/essential/2026-09-11.md), because
	// answering false silently PERMITS a UoM edit that should have been blocked. Changing it
	// needs a decision from the owner of essential's guard.
	engineRepo, err := services.RepositoryFor(schemaName)
	if err != nil {
		return false, nil
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
		// Archived lines still count: an archived order's lines are history, which is what must not
		// be reinterpreted.
		IncludeArchived: nil,
	})
	if err != nil {
		return false, errors.Wrapf(err, "checking whether %s references %s", schemaName, field)
	}
	return found != nil && found.HasData && len(found.Data.Items) > 0, nil
}
