package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// BulkCreateCommand is the wire form of POST /{resource}/bulk: {org_id?, items: [{...}, ...]}.
type BulkCreateCommand = dmodel.DynamicFields

const fieldNameItems = "items"

// Row-level error codes shared by bulk create and import. They are translation keys.
const (
	ErrRowValidation          = "err_import_validation"
	ErrRowDuplicateExternalId = "err_import_duplicate_external_id"
	ErrRowExternalIdOtherOrg  = "err_import_external_id_in_other_org"
)

// BulkRow is one record to write, tagged with its 1-based position so a rejection can be
// pointed back at the caller's list or spreadsheet.
type BulkRow struct {
	Number int
	Fields dmodel.DynamicFields
}

// paramsToBulkRows reads "items" out of the command. Shape problems are client errors: the
// caller sent a malformed request rather than bad data.
func paramsToBulkRows(cmd BulkCreateCommand) ([]BulkRow, *ft.ClientErrors) {
	raw, ok := cmd[fieldNameItems]
	if !ok || raw == nil {
		return nil, singleClientError(ft.NewValidationError(fieldNameItems, "err_import_items_required", "items are required"))
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, singleClientError(ft.NewValidationError(fieldNameItems, "err_import_items_not_a_list", "items must be a list"))
	}
	rows := make([]BulkRow, 0, len(list))
	for i, item := range list {
		fields, ok := item.(map[string]any)
		if !ok {
			return nil, singleClientError(ft.NewValidationError(
				fieldNameItems, "err_import_item_not_a_record", "every item must be an object",
				map[string]any{"index": i},
			))
		}
		rows = append(rows, BulkRow{Number: i + 1, Fields: dmodel.DynamicFields(fields)})
	}
	return rows, nil
}

func singleClientError(item *ft.ClientErrorItem) *ft.ClientErrors {
	cErrs := ft.NewClientErrors()
	cErrs.Append(*item)
	return cErrs
}

// SchemaSupportsDedup reports whether a schema carries the two deduplication fields, which is
// what switches bulk create from "insert every row" to "insert or update by external key".
func SchemaSupportsDedup(schema *dmodel.ModelSchema) bool {
	if schema == nil {
		return false
	}
	_, hasSource := schema.Field(FieldSourceSystem)
	_, hasExternal := schema.Field(FieldExternalId)
	return hasSource && hasExternal
}

// RunBulkCreate writes rows through domSvc inside one transaction and reports the outcome.
//
// domSvc must be the outermost domain service of the resource (the module's own, when it has
// one) so that its Create/Update overrides and hooks run for every row exactly as they would
// for a single call. A row the schema or a hook rejects is reported in the result and skipped;
// a Go error - including a database refusal such as the unique-key last defence - aborts and
// rolls back every row.
//
// rejected are rows the caller already refused (import conversion, unresolved references);
// they count toward TotalRows and are echoed in the errors list.
func RunBulkCreate(
	ctx corectx.Context, domSvc CrudDomainService, rows []BulkRow, rejected []RowError,
) (*BulkCreateResult, error) {
	return runBulkCreate(ctx, domSvc, func(_ corectx.Context) ([]BulkRow, []RowError, error) {
		return rows, rejected, nil
	})
}

// prepareRowsFn produces the rows to write from inside the transaction, so a preparation that
// itself writes (import creating referenced records) rolls back together with the rows.
type prepareRowsFn func(tranxCtx corectx.Context) (rows []BulkRow, rejected []RowError, err error)

func runBulkCreate(ctx corectx.Context, domSvc CrudDomainService, prepare prepareRowsFn) (*BulkCreateResult, error) {
	writer := &bulkWriter{
		domSvc: domSvc,
		schema: domSvc.Schema(),
		dedup:  SchemaSupportsDedup(domSvc.Schema()),
	}
	var data BulkCreateResultData
	err := WithTransaction(ctx, domSvc.Repository(), func(tranxCtx corectx.Context) error {
		rows, rejected, err := prepare(tranxCtx)
		if err != nil {
			return err
		}
		data.TotalRows = len(rows) + len(rejected)
		data.Errors = append([]RowError{}, rejected...)
		if writer.dedup {
			rows, data.Errors = writer.dropInFileDuplicates(rows, data.Errors)
			if writer.existing, err = writer.findExisting(tranxCtx, rows); err != nil {
				return err
			}
		}
		return writer.writeRows(tranxCtx, rows, &data)
	})
	if signal, ok := AsClientErrorSignal(err); ok {
		return &BulkCreateResult{ClientErrors: signal.Errors}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "RunBulkCreate")
	}

	sortRowErrors(data.Errors)
	data.AffectedCount = data.CreatedCount + data.UpdatedCount
	data.AffectedAt = model.NewModelDateTime()
	data.ErrorCount = len(data.Errors)
	return &BulkCreateResult{Data: data, HasData: true}, nil
}

func (this *bulkWriter) writeRows(ctx corectx.Context, rows []BulkRow, data *BulkCreateResultData) error {
	for _, row := range rows {
		outcome, rowErr, err := this.writeRow(ctx, row)
		if err != nil {
			return err
		}
		if rowErr != nil {
			data.Errors = append(data.Errors, *rowErr)
			continue
		}
		if outcome == rowUpdated {
			data.UpdatedCount++
		} else {
			data.CreatedCount++
		}
	}
	return nil
}

type rowOutcome int

const (
	rowCreated rowOutcome = iota
	rowUpdated
)

// existingRecord is what the dedup lookup learned about a stored row.
type existingRecord struct {
	id    string
	orgId string
	etag  string
}

type bulkWriter struct {
	domSvc   CrudDomainService
	schema   *dmodel.ModelSchema
	dedup    bool
	existing map[string]existingRecord
}

func (this *bulkWriter) writeRow(ctx corectx.Context, row BulkRow) (rowOutcome, *RowError, error) {
	if found, ok := this.existingFor(row.Fields); ok {
		if rowErr := this.assertSameOrg(row, found); rowErr != nil {
			return rowUpdated, rowErr, nil
		}
		cmd := copyFields(row.Fields)
		cmd[basemodel.FieldId] = found.id
		if found.etag != "" {
			cmd[basemodel.FieldEtag] = found.etag
		}
		result, err := this.domSvc.Update(ctx, cmd)
		if err != nil {
			return rowUpdated, nil, err
		}
		return rowUpdated, rowErrorFrom(row, result.ClientErrors), nil
	}

	result, err := this.domSvc.Create(ctx, copyFields(row.Fields))
	if err != nil {
		return rowCreated, nil, err
	}
	return rowCreated, rowErrorFrom(row, result.ClientErrors), nil
}

// assertSameOrg refuses to update a record that belongs to another organization: the external
// key is unique per tenant, not per org, so a collision across orgs is a data error, not an
// update the importing org is entitled to.
func (this *bulkWriter) assertSameOrg(row BulkRow, found existingRecord) *RowError {
	if _, hasOrg := this.schema.Field(basemodel.FieldOrgId); !hasOrg {
		return nil
	}
	rowOrg := readString(row.Fields, basemodel.FieldOrgId)
	if rowOrg == "" || found.orgId == "" || rowOrg == found.orgId {
		return nil
	}
	return &RowError{
		Row:   row.Number,
		Field: FieldExternalId,
		Code:  ErrRowExternalIdOtherOrg,
		Params: map[string]any{
			FieldExternalId: readString(row.Fields, FieldExternalId),
		},
	}
}

func rowErrorFrom(row BulkRow, cErrs ft.ClientErrors) *RowError {
	if cErrs.Count() == 0 {
		return nil
	}
	field := ""
	if len(cErrs) == 1 {
		field = cErrs[0].Field
	}
	return &RowError{
		Row:    row.Number,
		Field:  field,
		Code:   ErrRowValidation,
		Params: map[string]any{"errors": cErrs},
	}
}

func copyFields(fields dmodel.DynamicFields) dmodel.DynamicFields {
	out := make(dmodel.DynamicFields, len(fields)+1)
	for key, val := range fields {
		out[key] = val
	}
	return out
}
