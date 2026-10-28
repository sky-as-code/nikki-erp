package app

import (
	"sort"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
)

// attachEdgeLabels names the records an edge field points at, so the history reads
// "Role A, Role B (Roles)" rather than raw ids. The ids of one page are gathered per target schema
// and each schema is read once, as the caller: a target they may not read stays unlabelled.
func (this *ActionApplicationServiceImpl) attachEdgeLabels(
	ctx corectx.Context, target resourceaccess.ReadAsCallerQuery, items []itAction.HistoryItem,
) error {
	edgeTargets := edgeTargetsOf(dmodel.GetSchema(target.ResourceCode), items)
	if len(edgeTargets) == 0 {
		return nil
	}
	idsBySchema := collectEdgeIds(edgeTargets, items)
	labelsBySchema := map[string]map[string]string{}
	for _, destSchema := range sortedKeys(idsBySchema) {
		labels, err := this.access.LabelsAsCaller(ctx, resourceaccess.LabelsQuery{
			ResourceCode: destSchema, Ids: idsBySchema[destSchema], OrgId: target.OrgId,
		})
		if err != nil {
			return err
		}
		labelsBySchema[destSchema] = labels
	}
	applyEdgeLabels(edgeTargets, labelsBySchema, items)
	return nil
}

// edgeTargetsOf maps each changed field that is an edge to the schema it points at: a
// many-to-many edge by its name (or, for an entry written before edges were named, by the peer
// schema name), a to-one edge by its foreign-key column.
func edgeTargetsOf(schema *dmodel.ModelSchema, items []itAction.HistoryItem) map[string]string {
	targets := map[string]string{}
	if schema == nil {
		return targets
	}
	for _, item := range items {
		for _, change := range item.Changes {
			if _, known := targets[change.Field]; known {
				continue
			}
			if dest := edgeTargetOf(schema, change.Field); dest != "" {
				targets[change.Field] = dest
			}
		}
	}
	return targets
}

func edgeTargetOf(schema *dmodel.ModelSchema, field string) string {
	if link, ok := schema.M2mPeerLinkForEdge(field); ok && link.DestSchema != nil {
		return link.DestSchema.Name()
	}
	if link, ok := schema.M2mPeerLinkForDest(field); ok && link.DestSchema != nil {
		return link.DestSchema.Name()
	}
	for _, relation := range schema.ToRelations() {
		if relation.IsInverse || relation.RelationType == dmodel.RelationTypeManyToMany ||
			relation.RelationType == dmodel.RelationTypeOneToMany {
			continue
		}
		keys := relation.EffectiveForeignKeys()
		if len(keys) == 1 && keys[0].FkColumn == field {
			return relation.DestSchemaName
		}
	}
	return ""
}

func collectEdgeIds(edgeTargets map[string]string, items []itAction.HistoryItem) map[string][]model.Id {
	seen := map[string]map[string]bool{}
	for _, item := range items {
		for _, change := range item.Changes {
			dest, isEdge := edgeTargets[change.Field]
			if !isEdge {
				continue
			}
			if seen[dest] == nil {
				seen[dest] = map[string]bool{}
			}
			for _, id := range append(idsOf(change.Old), idsOf(change.New)...) {
				seen[dest][id] = true
			}
		}
	}
	out := map[string][]model.Id{}
	for dest, ids := range seen {
		for _, id := range sortedKeys(ids) {
			out[dest] = append(out[dest], model.Id(id))
		}
	}
	return out
}

func applyEdgeLabels(
	edgeTargets map[string]string, labelsBySchema map[string]map[string]string, items []itAction.HistoryItem,
) {
	for i := range items {
		for j := range items[i].Changes {
			change := &items[i].Changes[j]
			dest, isEdge := edgeTargets[change.Field]
			if !isEdge {
				continue
			}
			labels := map[string]string{}
			for _, id := range append(idsOf(change.Old), idsOf(change.New)...) {
				if label, ok := labelsBySchema[dest][id]; ok {
					labels[id] = label
				}
			}
			if len(labels) > 0 {
				change.Labels = labels
			}
		}
	}
}

// idsOf reads the ids a change value holds: one id for a foreign key, an array for an m2m edge.
func idsOf(value any) []string {
	switch typed := value.(type) {
	case string:
		if typed != "" {
			return []string{typed}
		}
	case []any:
		ids := []string{}
		for _, item := range typed {
			if id, ok := item.(string); ok && id != "" {
				ids = append(ids, id)
			}
		}
		return ids
	case []string:
		return typed
	}
	return nil
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
