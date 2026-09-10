package composable

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// dedupLookupChunk bounds the IN list of one existing-key query.
const dedupLookupChunk = 500

// dedupKey is the in-memory form of the (source_system, external_id) unique key.
func dedupKey(sourceSystem string, externalId string) string {
	return sourceSystem + "\x00" + externalId
}

// dedupKeyOf reads the row's key, filling a missing source_system with the schema default so a
// row that relies on the default still matches the stored record it will be written as.
func (this *bulkWriter) dedupKeyOf(fields dmodel.DynamicFields) (string, bool) {
	externalId := readString(fields, FieldExternalId)
	if externalId == "" {
		return "", false
	}
	return dedupKey(this.sourceSystemOf(fields), externalId), true
}

func (this *bulkWriter) sourceSystemOf(fields dmodel.DynamicFields) string {
	if source := readString(fields, FieldSourceSystem); source != "" {
		return source
	}
	if field, ok := this.schema.Field(FieldSourceSystem); ok {
		if def := field.Default(); def != nil && !def.IsEmpty() {
			if str, ok := (*def.Get()).(string); ok {
				return str
			}
		}
	}
	return SourceSystemManual
}

// dropInFileDuplicates keeps the first row of every external key and reports the rest: two rows
// naming one record would otherwise race each other inside the transaction.
func (this *bulkWriter) dropInFileDuplicates(rows []BulkRow, rowErrors []RowError) ([]BulkRow, []RowError) {
	seen := make(map[string]int, len(rows))
	kept := make([]BulkRow, 0, len(rows))
	for _, row := range rows {
		key, ok := this.dedupKeyOf(row.Fields)
		if !ok {
			kept = append(kept, row)
			continue
		}
		if first, dup := seen[key]; dup {
			rowErrors = append(rowErrors, RowError{
				Row:   row.Number,
				Field: FieldExternalId,
				Code:  ErrRowDuplicateExternalId,
				Params: map[string]any{
					FieldExternalId: readString(row.Fields, FieldExternalId),
					"first_row":     first,
				},
			})
			continue
		}
		seen[key] = row.Number
		kept = append(kept, row)
	}
	return kept, rowErrors
}

// findExisting fetches, in chunks, every stored record whose external_id appears in rows and
// indexes it by the full key. It reads through the resource's own repository so tenant and
// archive handling stay what a direct read would get.
func (this *bulkWriter) findExisting(ctx corectx.Context, rows []BulkRow) (map[string]existingRecord, error) {
	externalIds := distinctExternalIds(rows)
	found := make(map[string]existingRecord, len(externalIds))
	fields := []string{basemodel.FieldId, FieldSourceSystem, FieldExternalId}
	if _, hasOrg := this.schema.Field(basemodel.FieldOrgId); hasOrg {
		fields = append(fields, basemodel.FieldOrgId)
	}
	// A versioned schema refuses an update that does not carry the stored etag.
	if _, hasEtag := this.schema.Field(basemodel.FieldEtag); hasEtag {
		fields = append(fields, basemodel.FieldEtag)
	}
	for start := 0; start < len(externalIds); start += dedupLookupChunk {
		end := min(start+dedupLookupChunk, len(externalIds))
		items, err := SearchRepositoryRows(
			ctx, this.domSvc.Repository(), this.schema.Name(), FieldExternalId, externalIds[start:end], fields,
		)
		if err != nil {
			return nil, errors.Wrap(err, "bulk create dedup lookup")
		}
		for _, item := range items {
			key := dedupKey(readString(item, FieldSourceSystem), readString(item, FieldExternalId))
			found[key] = existingRecord{
				id:    readString(item, basemodel.FieldId),
				orgId: readString(item, basemodel.FieldOrgId),
				etag:  readString(item, basemodel.FieldEtag),
			}
		}
	}
	return found, nil
}

func (this *bulkWriter) existingFor(fields dmodel.DynamicFields) (existingRecord, bool) {
	if !this.dedup {
		return existingRecord{}, false
	}
	key, ok := this.dedupKeyOf(fields)
	if !ok {
		return existingRecord{}, false
	}
	found, ok := this.existing[key]
	return found, ok
}

func distinctExternalIds(rows []BulkRow) []any {
	seen := map[string]bool{}
	ids := make([]any, 0, len(rows))
	for _, row := range rows {
		externalId := readString(row.Fields, FieldExternalId)
		if externalId == "" || seen[externalId] {
			continue
		}
		seen[externalId] = true
		ids = append(ids, externalId)
	}
	return ids
}

func sortRowErrors(rowErrors []RowError) {
	sort.SliceStable(rowErrors, func(i, j int) bool {
		return rowErrors[i].Row < rowErrors[j].Row
	})
}
