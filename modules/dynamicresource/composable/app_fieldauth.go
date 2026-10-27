package composable

import (
	"encoding/json"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// Field-level authorization of the default application service, exported through
// CrudApplicationService so a custom action applies the same rules. See requestguard's
// field_permission.go for the grant semantics.

// AssertFieldAction is AssertAction for an action whose reach can be narrowed per field (create,
// read, update): the caller passes with a grant on the whole resource or on at least one of its
// fields. Pair it with AssertFields on what is written, or NarrowReadFields on what is returned -
// on its own it lets a holder of one field grant through.
func (this *DefaultApplicationServiceImpl) AssertFieldAction(
	ctx corectx.Context, actionCode string, params dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	orgId, cErrs := this.ResolveOrgScope(ctx, params)
	if cErrs != nil {
		return nil, cErrs
	}
	return orgId, this.assertAnyField(ctx, actionCode, orgId)
}

func (this *DefaultApplicationServiceImpl) assertAnyField(
	ctx corectx.Context, actionCode string, orgId *model.Id,
) *ft.ClientErrors {
	return this.assertAnyFieldIn(ctx, actionCode, recordAccess{orgId: orgId})
}

func (this *DefaultApplicationServiceImpl) assertAnyFieldIn(
	ctx corectx.Context, actionCode string, access recordAccess,
) *ft.ClientErrors {
	return requestguard.AssertAnyFieldPermission(ctx, this.permFor(actionCode, access), this.schemaFieldNames())
}

// AssertFields authorizes actionCode on every schema field fields carries. Keys naming no schema
// field (paging, "fields", a virtual param) and the addressing keys (see isAddressingField) are
// not data the caller writes or reads, so they need no field grant.
func (this *DefaultApplicationServiceImpl) AssertFields(
	ctx corectx.Context, actionCode string, orgId *model.Id, fields dmodel.DynamicFields,
) *ft.ClientErrors {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	return this.AssertFieldNames(ctx, actionCode, orgId, names)
}

func (this *DefaultApplicationServiceImpl) assertFieldsIn(
	ctx corectx.Context, actionCode string, access recordAccess, fields dmodel.DynamicFields,
) *ft.ClientErrors {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	return this.assertFieldNamesIn(ctx, actionCode, access, names)
}

// AssertFieldNames is AssertFields over field names, for a payload that is not a field map.
func (this *DefaultApplicationServiceImpl) AssertFieldNames(
	ctx corectx.Context, actionCode string, orgId *model.Id, names []string,
) *ft.ClientErrors {
	return this.assertFieldNamesIn(ctx, actionCode, recordAccess{orgId: orgId}, names)
}

func (this *DefaultApplicationServiceImpl) assertFieldNamesIn(
	ctx corectx.Context, actionCode string, access recordAccess, names []string,
) *ft.ClientErrors {
	return requestguard.AssertFieldNamesPermission(ctx, this.permFor(actionCode, access), this.guardedFields(names))
}

// NarrowReadFields blanks, in place, every field of records the caller may not read, and returns
// the blanked names for the response's masked_fields. Following the masked-field contract of
// dyn.PagedResultData, a blanked field stays present as null, so a client can tell "locked" from
// "absent". desiredFields widens the masked list to fields asked for but missing from every record.
func (this *DefaultApplicationServiceImpl) NarrowReadFields(
	ctx corectx.Context, orgId *model.Id, desiredFields []string, records ...dmodel.DynamicFields,
) []string {
	return this.narrowIn(ctx, recordAccess{orgId: orgId}, desiredFields, records...)
}

func (this *DefaultApplicationServiceImpl) narrowIn(
	ctx corectx.Context, access recordAccess, desiredFields []string, records ...dmodel.DynamicFields,
) []string {
	perm := this.permFor(PermissionRead, access)
	if requestguard.HasPermission(ctx, perm) {
		return nil
	}
	candidates := this.guardedFields(append(topLevelNames(desiredFields), recordKeys(records)...))
	readable := toSet(requestguard.ReadableFields(ctx, perm, candidates))
	masked := make([]string, 0)
	for _, name := range candidates {
		if _, ok := readable[name]; ok {
			continue
		}
		masked = append(masked, name)
		for _, record := range records {
			if _, present := record[name]; present {
				record[name] = nil
			}
		}
	}
	return masked
}

// AssertGraphReadable refuses a search whose graph filters or sorts on a field the caller may not
// read: the rows it returns would disclose the field's value even with the column masked.
func (this *DefaultApplicationServiceImpl) AssertGraphReadable(
	ctx corectx.Context, orgId *model.Id, rawGraph any,
) (*ft.ClientErrors, error) {
	return this.assertGraphReadableIn(ctx, recordAccess{orgId: orgId}, rawGraph)
}

func (this *DefaultApplicationServiceImpl) assertGraphReadableIn(
	ctx corectx.Context, access recordAccess, rawGraph any,
) (*ft.ClientErrors, error) {
	if rawGraph == nil || requestguard.HasPermission(ctx, this.permFor(PermissionRead, access)) {
		return nil, nil
	}
	graph, err := decodeSearchGraph(rawGraph)
	if err != nil {
		return nil, err
	}
	names := graphFieldNames(graph.ToSearchNode(), nil)
	for _, item := range graph.GetOrder() {
		if len(item) > 0 {
			names = append(names, item.Field())
		}
	}
	return this.assertFieldNamesIn(ctx, PermissionRead, access, topLevelNames(names)), nil
}

// narrowSingle masks the unreadable fields of a get-one result, for chaining onto the domain call.
func (this *DefaultApplicationServiceImpl) narrowSingle(
	ctx corectx.Context, access recordAccess, mask edgeMask,
) func(*GetOneResult, error) (*GetOneResult, error) {
	return func(result *GetOneResult, err error) (*GetOneResult, error) {
		if err != nil || result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
			return result, err
		}
		meta := &result.Data.Meta
		mask.applyTo(&meta.DesiredFields, &meta.MaskedFields, result.Data.Item)
		masked := this.narrowIn(ctx, access, meta.DesiredFields, result.Data.Item)
		meta.MaskedFields = appendMissing(meta.MaskedFields, masked)
		return result, nil
	}
}

// narrowPaged is narrowSingle for a search result.
func (this *DefaultApplicationServiceImpl) narrowPaged(
	ctx corectx.Context, access recordAccess, mask edgeMask,
) func(*SearchResult, error) (*SearchResult, error) {
	return func(result *SearchResult, err error) (*SearchResult, error) {
		if err != nil || result == nil || result.ClientErrors.Count() > 0 {
			return result, err
		}
		data := &result.Data
		mask.applyTo(&data.DesiredFields, &data.MaskedFields, data.Items...)
		masked := this.narrowIn(ctx, access, data.DesiredFields, data.Items...)
		data.MaskedFields = appendMissing(data.MaskedFields, masked)
		return result, nil
	}
}

func searchFailure(cErrs *ft.ClientErrors, err error) (*SearchResult, error) {
	if err != nil {
		return nil, err
	}
	return &SearchResult{ClientErrors: *cErrs}, nil
}

func appendMissing(target []string, items []string) []string {
	present := toSet(target)
	for _, item := range items {
		if _, ok := present[item]; !ok {
			target = append(target, item)
			present[item] = struct{}{}
		}
	}
	return target
}

func (this *DefaultApplicationServiceImpl) permFor(actionCode string, access recordAccess) requestguard.Perm {
	return requestguard.PermFor(actionCode, this.ResourceCode(), this.scope).
		InOrg(access.orgId).OwnedByCaller(access.owned)
}

func (this *DefaultApplicationServiceImpl) schemaFieldNames() []string {
	schema := this.Schema()
	if schema == nil {
		return nil
	}
	return this.guardedFields(schema.FieldNames())
}

// guardedFields keeps the names that need a field grant: schema fields other than the addressing
// ones.
func (this *DefaultApplicationServiceImpl) guardedFields(names []string) []string {
	schema := this.Schema()
	if schema == nil {
		return nil
	}
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		if field, exists := schema.Field(name); exists && !isAddressingField(field) {
			result = append(result, name)
		}
	}
	return result
}

// isAddressingField reports the keys that locate a record rather than describe it: the primary
// key, the etag, the tenant key and org_id. They are readable and writable under any grant on the
// action - a holder of one field grant still needs them to find the record and to save it - and
// the org itself is already checked by ResolveOrgScope.
func isAddressingField(field *dmodel.ModelField) bool {
	switch field.Name() {
	case basemodel.FieldId, basemodel.FieldEtag, basemodel.FieldOrgId:
		return true
	}
	return field.IsPrimaryKey() ||
		field.IsVersioningKey() || field.IsTenantKey()
}

// topLevelNames reduces "edge.field" paths to the field of this resource they go through.
func topLevelNames(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		name, _, _ := strings.Cut(path, ".")
		result = append(result, name)
	}
	return result
}

func recordKeys(records []dmodel.DynamicFields) []string {
	result := make([]string, 0)
	for _, record := range records {
		for key := range record {
			result = append(result, key)
		}
	}
	return result
}

func toSet(items []string) map[string]struct{} {
	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		result[item] = struct{}{}
	}
	return result
}

func decodeSearchGraph(raw any) (*dmodel.SearchGraph, error) {
	if graph, ok := raw.(*dmodel.SearchGraph); ok {
		return graph, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, errors.Wrap(err, "decodeSearchGraph.Marshal")
	}
	graph := dmodel.NewSearchGraph()
	if err := json.Unmarshal(encoded, graph); err != nil {
		return nil, errors.Wrap(err, "decodeSearchGraph.Unmarshal")
	}
	return graph, nil
}

func graphFieldNames(node *dmodel.SearchNode, names []string) []string {
	if node == nil {
		return names
	}
	if condition := node.GetCondition(); len(condition) > 0 {
		names = append(names, condition.Field())
	}
	for _, child := range node.GetAnd() {
		names = graphFieldNames(&child, names)
	}
	for _, child := range node.GetOr() {
		names = graphFieldNames(&child, names)
	}
	return names
}
