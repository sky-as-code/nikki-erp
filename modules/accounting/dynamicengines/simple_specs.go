package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// The resources whose every invariant the schema already expresses.
//
// A spec with no DefineActions is not an oversight: the dynamic resource engine already enforces
// required fields, types, bounds, enum membership and the unique indexes declared in the JSON. Only
// rules the schema cannot state - a cycle, a conditional requirement, an overlap between two rows -
// need code, and those live in their own files.

func taxGroupEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.TaxGroupSchemaName}
}

func taxProductClassificationEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.TaxProductClassificationSchemaName}
}

func taxEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.TaxSchemaName}
}

func taxMappingLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxMappingLineSchemaName,
		DefineActions: defineMappingLineActions,
	}
}
