package composable

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// Edge masking: an edge, or an `edge.leaf` path, named in `fields=` that the caller may not read is
// taken out of the projection before the query runs, so its data is never fetched. The response
// lists the requested name in masked_fields and, when nothing of the edge is left, carries the edge
// as null - the masked-field contract of a scalar, applied to an edge. Projection never answers 403;
// filtering or sorting on an unreadable field still does (AssertGraphReadable).

// edgeMask is what stripUnreadableEdges took out of a request, for applyTo to report afterwards.
type edgeMask struct {
	// requested are the names as the caller wrote them in fields=.
	requested []string
	// nulledEdges are the edges with no readable path left; their value is answered as null.
	nulledEdges []string
}

func (this edgeMask) isEmpty() bool {
	return len(this.requested) == 0
}

// applyTo writes the mask into a read result: every stripped name is desired and masked, and every
// fully stripped edge is present as null in each record.
func (this edgeMask) applyTo(desired, masked *[]string, records ...dmodel.DynamicFields) {
	if this.isEmpty() {
		return
	}
	*desired = appendMissing(*desired, this.requested)
	*masked = appendMissing(*masked, this.requested)
	for _, record := range records {
		for _, edge := range this.nulledEdges {
			record[edge] = nil
		}
	}
}

// stripUnreadableEdges removes from params' fields every edge path the caller may not read and
// returns what it removed. Only an explicit fields= is inspected: the default projections are
// scalar columns.
func (this *DefaultApplicationServiceImpl) stripUnreadableEdges(
	ctx corectx.Context, access recordAccess, params dmodel.DynamicFields,
) edgeMask {
	schema := this.Schema()
	fields := requestedFields(params)
	if schema == nil || len(fields) == 0 {
		return edgeMask{}
	}
	kept := make([]string, 0, len(fields))
	mask := edgeMask{}
	keptEdges := make(map[string]bool)
	for _, path := range fields {
		edge, isEdge := this.edgeOf(schema, path)
		if !isEdge || this.canReadEdgePath(ctx, access, edge, path) {
			kept = append(kept, path)
			keptEdges[edge.Edge] = keptEdges[edge.Edge] || isEdge
			continue
		}
		mask.requested = append(mask.requested, path)
		if !array.Contains(mask.nulledEdges, edge.Edge) {
			mask.nulledEdges = append(mask.nulledEdges, edge.Edge)
		}
	}
	if mask.isEmpty() {
		return mask
	}
	mask.nulledEdges = array.Filter(mask.nulledEdges, func(edge string) bool { return !keptEdges[edge] })
	if len(kept) == 0 {
		// An empty selection would fall back to the default projection, returning what was not asked.
		kept = keyFields(schema)
	}
	params[fieldNameFields] = kept
	return mask
}

func (this *DefaultApplicationServiceImpl) edgeOf(schema *dmodel.ModelSchema, path string) (dmodel.ModelRelation, bool) {
	root, _, _ := strings.Cut(path, ".")
	for _, relation := range schema.Relations() {
		if relation.Edge == root {
			return relation, true
		}
	}
	return dmodel.ModelRelation{}, false
}

// canReadEdgePath answers whether the caller may read path through edge: the edge field of this
// resource must be readable (resource or field grant), and so must the destination - the whole
// destination resource for a bare edge or a wildcard, the resource or that leaf field otherwise.
// The destination is checked at this resource's scope and org: the edge's records are reached
// through this record, within the org the request is confined to.
func (this *DefaultApplicationServiceImpl) canReadEdgePath(
	ctx corectx.Context, access recordAccess, edge dmodel.ModelRelation, path string,
) bool {
	sourcePerm := this.permFor(PermissionRead, access)
	if !requestguard.HasPermission(ctx, sourcePerm) &&
		!requestguard.HasPermission(ctx, sourcePerm.ForField(edge.Edge)) {
		return false
	}
	destPerm := requestguard.PermFor(PermissionRead, edge.DestSchemaName, this.scope).InOrg(access.orgId)
	if requestguard.HasPermission(ctx, destPerm) {
		return true
	}
	_, rest, hasLeaf := strings.Cut(path, ".")
	leaf, _, _ := strings.Cut(rest, ".")
	if !hasLeaf || leaf == "" || leaf == corecrud.FieldsWildcard {
		return false
	}
	return requestguard.HasPermission(ctx, destPerm.ForField(leaf))
}

func keyFields(schema *dmodel.ModelSchema) []string {
	if keys := schema.PrimaryKeys(); len(keys) > 0 {
		return keys
	}
	return []string{basemodel.FieldId}
}

// requestedFields reads fields= as given, nil when the caller asked for no explicit selection.
func requestedFields(params dmodel.DynamicFields) []string {
	switch raw := params[fieldNameFields].(type) {
	case []string:
		return raw
	case []any:
		result := make([]string, 0, len(raw))
		for _, item := range raw {
			if name, ok := item.(string); ok {
				result = append(result, name)
			}
		}
		return result
	}
	return nil
}
