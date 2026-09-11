package composable

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/tabular"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// importRows is the domain half of an import: it binds the mapping to the file, types every
// cell, resolves reference labels inside the write transaction and hands the rows to bulk
// create. It takes the outermost domain service for the same reason RunBulkCreate does.
func importRows(
	ctx corectx.Context, domSvc CrudDomainService, table *tabular.Table, mapping ImportMapping,
	orgId *model.Id, lookupOnion onionLookupFn,
) (*BulkCreateResult, error) {
	schema := domSvc.Schema()
	plan, cErrs := planImport(schema, table.Headers, mapping)
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	resolver := &referenceResolver{
		plan:          plan,
		createMissing: mapping.CreateMissingReferences,
		orgId:         orgIdString(orgId),
		lookupOnion:   lookupOnion,
	}
	targets, cErrs := resolver.assertTargetsServed()
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}

	rows, rejected := plan.convertRows(table.Rows)
	forceOrgId(rows, orgId)
	stampImportSource(schema, rows)

	return runBulkCreate(ctx, domSvc, func(tranxCtx corectx.Context) ([]BulkRow, []RowError, error) {
		kept, referenceErrors, err := resolver.resolve(tranxCtx, targets, rows)
		if err != nil {
			return nil, nil, err
		}
		return kept, append(rejected, referenceErrors...), nil
	})
}

// stampImportSource marks a row that names no source system as coming from an import, so its
// external id is recognised again by the next import rather than mistaken for a manual entry.
func stampImportSource(schema *dmodel.ModelSchema, rows []BulkRow) {
	if !SchemaSupportsDedup(schema) {
		return
	}
	for _, row := range rows {
		if readString(row.Fields, FieldSourceSystem) == "" {
			row.Fields[FieldSourceSystem] = SourceSystemImport
		}
	}
}

func orgIdString(orgId *model.Id) string {
	if orgId == nil {
		return ""
	}
	return string(*orgId)
}
