package cqrs

import (
	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
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

func uomTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.TaxRateVersionSchemaName, Field: models.TaxRateVersionFieldRateUomId},
	}
}

// resolveRepository reaches Accounting's own repositories through the typed ports the tax
// configuration onions publish, at check time rather than at registration, so subscribing the
// handler during Init builds no onion. A schema this probe does not know is an error: an unanswered
// check refuses the delete, which is the safe direction.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	if schemaName != models.TaxRateVersionSchemaName {
		return nil, errors.Errorf("no repository for %s", schemaName)
	}
	var found usagecheck.RowSearcher
	err := deps.Invoke(func(rateVersions taxconfig.TaxRateVersionRepository) {
		found = rateVersions
	})
	if err != nil {
		return nil, errors.Wrapf(err, "no repository for %s", schemaName)
	}
	return found, nil
}
