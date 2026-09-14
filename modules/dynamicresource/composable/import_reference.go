package composable

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// referenceLookupChunk bounds the IN list of one label query.
const referenceLookupChunk = 500

// onionLookupFn finds the onion serving a schema; LookupSourceOnion in production, a fake in tests.
type onionLookupFn func(schemaName string) (DynamicResourceEngineOnion, bool)

// referenceResolver replaces the label a reference cell carries with the id of the referenced
// record. Labels are looked up per referenced schema in batches; a label nobody stores is a
// row error unless the caller allowed creating it, in which case the referenced resource's own
// domain service creates it so its rules run.
type referenceResolver struct {
	plan          *importPlan
	createMissing bool
	orgId         string
	lookupOnion   onionLookupFn
}

// referenceTarget is one referenced schema with everything resolving against it needs.
type referenceTarget struct {
	field       string
	onion       DynamicResourceEngineOnion
	schema      *dmodel.ModelSchema
	labelField  string
	labelIsLang bool
	hasOrg      bool
}

// assertTargetsServed checks, before any file row is read, that every referenced schema is
// served by a composable onion and declares a record label. Both are configuration gaps the
// caller cannot fix, but they are reported as client errors so the mapping page can name them.
func (this *referenceResolver) assertTargetsServed() ([]referenceTarget, *ft.ClientErrors) {
	targets := make([]referenceTarget, 0)
	cErrs := ft.NewClientErrors()
	for _, column := range this.plan.columns {
		if column.relation == nil {
			continue
		}
		onion, ok := this.lookupOnion(column.relation.DestSchemaName)
		if !ok {
			cErrs.Append(*ft.NewValidationError(column.field.Name(), ErrImportMappingInvalid,
				"the referenced resource cannot be imported into",
				map[string]any{"schema": column.relation.DestSchemaName}))
			continue
		}
		schema := onion.Schema()
		labelField := schema.RecordLabelField()
		label, hasLabel := schema.Field(labelField)
		if labelField == "" || !hasLabel {
			cErrs.Append(*ft.NewValidationError(column.field.Name(), ErrImportMappingInvalid,
				"the referenced resource declares no record label",
				map[string]any{"schema": column.relation.DestSchemaName}))
			continue
		}
		if err := assertReferenceTargetLabel(column.field.Name(), schema, label); err != nil {
			cErrs.Append(*err)
			continue
		}
		_, hasOrg := schema.Field(basemodel.FieldOrgId)
		targets = append(targets, referenceTarget{
			field:       column.field.Name(),
			onion:       onion,
			schema:      schema,
			labelField:  labelField,
			labelIsLang: label.DataType().String() == dmodel.FieldDataTypeNameLangJson,
			hasOrg:      hasOrg,
		})
	}
	if cErrs.Count() > 0 {
		return nil, cErrs
	}
	return targets, nil
}

// assertReferenceTargetLabel refuses a referenced schema whose record label has no column.
//
// Resolving a reference means looking a record up *by* its label — `WHERE <label> IN (...)`,
// pushed to the database — and, when nothing matches, writing that label into a new record. A
// virtual label supports neither: it is filled after the read, so no query can filter on it and
// no create can set it. The model builder allows such a label because displaying one is fine;
// this is where the half that needs a column is enforced, at mapping time, so an import names
// the problem before it reads a single row rather than matching nothing at run time.
func assertReferenceTargetLabel(
	columnField string, schema *dmodel.ModelSchema, label *dmodel.ModelField,
) *ft.ClientErrorItem {
	if label.IsPersisted() {
		return nil
	}
	return ft.NewValidationError(columnField, ErrImportMappingInvalid,
		"the referenced resource's record label is computed, so records cannot be looked up by it",
		map[string]any{"schema": schema.Name(), "label_field": label.Name()})
}

// resolve rewrites every reference cell of rows in place and returns the rows that still
// qualify plus the errors of those that do not. It must run inside the import transaction so
// a referenced record created here rolls back with the rows that needed it.
func (this *referenceResolver) resolve(
	ctx corectx.Context, targets []referenceTarget, rows []BulkRow,
) ([]BulkRow, []RowError, error) {
	rejected := map[int]RowError{}
	for _, target := range targets {
		ids, err := this.resolveTarget(ctx, target, rows, rejected)
		if err != nil {
			return nil, nil, err
		}
		for _, row := range rows {
			if _, bad := rejected[row.Number]; bad {
				continue
			}
			label := readString(row.Fields, target.field)
			if label == "" {
				continue
			}
			row.Fields[target.field] = ids[normalizeLabel(label)]
		}
	}
	kept := make([]BulkRow, 0, len(rows))
	rowErrors := make([]RowError, 0, len(rejected))
	for _, row := range rows {
		if rowErr, bad := rejected[row.Number]; bad {
			rowErrors = append(rowErrors, rowErr)
			continue
		}
		kept = append(kept, row)
	}
	return kept, rowErrors, nil
}

// resolveTarget maps every distinct label of one reference column to an id, creating the
// missing ones when allowed, and records a row error for each row whose label stays unresolved.
func (this *referenceResolver) resolveTarget(
	ctx corectx.Context, target referenceTarget, rows []BulkRow, rejected map[int]RowError,
) (map[string]string, error) {
	labels := distinctLabels(rows, target.field)
	found, ambiguous, err := this.findByLabels(ctx, target, labels)
	if err != nil {
		return nil, err
	}
	failed := map[string]RowError{}
	for _, label := range labels {
		key := normalizeLabel(label)
		if _, ok := found[key]; ok {
			continue
		}
		if ambiguous[key] {
			failed[key] = RowError{Field: target.field, Code: ErrRowReferenceAmbiguous, Params: referenceParams(target, label)}
			continue
		}
		if !this.createMissing {
			failed[key] = RowError{Field: target.field, Code: ErrRowReferenceNotFound, Params: referenceParams(target, label)}
			continue
		}
		id, rowErr, err := this.createReferenced(ctx, target, label)
		if err != nil {
			return nil, err
		}
		if rowErr != nil {
			failed[key] = *rowErr
			continue
		}
		found[key] = id
	}
	for _, row := range rows {
		label := readString(row.Fields, target.field)
		if label == "" {
			continue
		}
		if rowErr, bad := failed[normalizeLabel(label)]; bad {
			if _, already := rejected[row.Number]; !already {
				rowErr.Row = row.Number
				rejected[row.Number] = rowErr
			}
		}
	}
	return found, nil
}

// findByLabels reads the referenced records whose label is one of labels, confined to the
// importing org when the referenced schema is org-scoped (a global record, with no org, still
// matches). Two records sharing a label make it ambiguous rather than picking one.
func (this *referenceResolver) findByLabels(
	ctx corectx.Context, target referenceTarget, labels []string,
) (found map[string]string, ambiguous map[string]bool, err error) {
	found = map[string]string{}
	ambiguous = map[string]bool{}
	fields := []string{basemodel.FieldId, target.labelField}
	if target.hasOrg {
		fields = append(fields, basemodel.FieldOrgId)
	}
	for start := 0; start < len(labels); start += referenceLookupChunk {
		end := min(start+referenceLookupChunk, len(labels))
		items, err := this.searchLabels(ctx, target, labels[start:end], fields)
		if err != nil {
			return nil, nil, err
		}
		for _, item := range items {
			if target.hasOrg && !this.inImportingOrg(item) {
				continue
			}
			key := normalizeLabel(this.labelOf(item, target))
			if _, dup := found[key]; dup {
				ambiguous[key] = true
				delete(found, key)
				continue
			}
			if !ambiguous[key] {
				found[key] = readString(item, basemodel.FieldId)
			}
		}
	}
	return found, ambiguous, nil
}

func (this *referenceResolver) searchLabels(
	ctx corectx.Context, target referenceTarget, labels []string, fields []string,
) ([]dmodel.DynamicFields, error) {
	keys := make([]any, len(labels))
	for i, label := range labels {
		keys[i] = label
	}
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(target.labelField, dmodel.In, keys...)
	param := dyn.RepoSearchParam{Fields: fields, Graph: graph, Page: 0, Size: len(keys)}
	if target.labelIsLang {
		language := this.plan.language
		param.Language = &language
	}
	result, err := target.onion.Repository().Search(ctx, param)
	if err != nil {
		return nil, errors.Wrapf(err, "import reference lookup on '%s'", target.schema.Name())
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return nil, errors.Errorf("import reference lookup on '%s' failed: %v",
			target.schema.Name(), result.ClientErrors.ToError())
	}
	if result == nil || !result.HasData {
		return nil, nil
	}
	return result.Data.Items, nil
}

func (this *referenceResolver) inImportingOrg(item dmodel.DynamicFields) bool {
	itemOrg := readString(item, basemodel.FieldOrgId)
	return itemOrg == "" || this.orgId == "" || itemOrg == this.orgId
}

// labelOf reads the stored label back in the mapping language, so a localized column compares
// against the same text the query matched on.
func (this *referenceResolver) labelOf(item dmodel.DynamicFields, target referenceTarget) string {
	raw := item[target.labelField]
	switch typed := raw.(type) {
	case string:
		return typed
	case model.LangJson:
		return typed[this.plan.language]
	case map[string]any:
		text, _ := typed[string(this.plan.language)].(string)
		return text
	case map[string]string:
		return typed[string(this.plan.language)]
	}
	return ""
}

// createReferenced writes the missing record with its label only (plus the importing org) through
// the referenced resource's own domain service. A refusal is the row's error, not the import's.
func (this *referenceResolver) createReferenced(
	ctx corectx.Context, target referenceTarget, label string,
) (string, *RowError, error) {
	cmd := dmodel.DynamicFields{}
	if target.labelIsLang {
		cmd[target.labelField] = model.LangJson{this.plan.language: label}
	} else {
		cmd[target.labelField] = label
	}
	if target.hasOrg && this.orgId != "" {
		cmd[basemodel.FieldOrgId] = this.orgId
	}
	result, err := target.onion.DomainService().Create(ctx, cmd)
	if err != nil {
		return "", nil, errors.Wrapf(err, "import create referenced '%s'", target.schema.Name())
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		params := referenceParams(target, label)
		params["errors"] = result.ClientErrors
		return "", &RowError{Field: target.field, Code: ErrRowReferenceCreateFailed, Params: params}, nil
	}
	id := readString(result.Data, basemodel.FieldId)
	if id == "" {
		return "", nil, errors.Errorf("import create referenced '%s' returned no id", target.schema.Name())
	}
	return id, nil, nil
}

func referenceParams(target referenceTarget, label string) map[string]any {
	return map[string]any{"schema": target.schema.Name(), "value": label}
}

func distinctLabels(rows []BulkRow, field string) []string {
	seen := map[string]bool{}
	labels := make([]string, 0)
	for _, row := range rows {
		label := readString(row.Fields, field)
		key := normalizeLabel(label)
		if label == "" || seen[key] {
			continue
		}
		seen[key] = true
		labels = append(labels, label)
	}
	return labels
}

// normalizeLabel is the in-memory match key: a label is the same record whatever its casing
// and surrounding whitespace, which is how a person reads a spreadsheet.
func normalizeLabel(label string) string {
	return strings.ToLower(strings.Join(strings.Fields(label), " "))
}
