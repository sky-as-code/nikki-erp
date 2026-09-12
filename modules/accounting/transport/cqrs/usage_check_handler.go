package cqrs

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// Accounting's answer to "is this unit of measure still in use", asked before Essential deletes
// or re-scales one.
//
// A specific-duty tax is charged per unit - so much per litre - and a rate version records both
// the amount and the unit it is per. Versions are effective-dated history that stays readable
// after being superseded, so a unit named by any version, current or not, is still in use:
// losing it would turn a recorded rate into a bare number.
//
// Accounting registered no probe before this, so a unit reachable only from tax rates looked
// unused.

func uomTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.TaxRateVersionSchemaName, Field: models.TaxRateVersionFieldRateUomId},
	}
}

// resolveRepository reaches Accounting's own repositories through the resource registry, at
// check time rather than at registration.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	engine, err := services.EngineFor(schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "no repository for %s", schemaName)
	}
	return engine.ResourceRepository(), nil
}
