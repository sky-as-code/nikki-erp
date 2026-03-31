package model

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
)

// FinalizeRelations runs all schema relation finalization after every model is registered:
// foreign-key map normalization, EdgeFrom peer resolution, then many-to-many junction wiring.
// Internal helpers must index reg.schemas directly instead of Get, because this method holds
// reg.mu.Lock and Get uses RLock (same goroutine would deadlock: RWMutex is not reentrant).
func (this *SchemaRegistry) FinalizeRelations() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	if err := normalizeAllForeignKeyMapsUnlocked(this); err != nil {
		return err
	}
	if err := finalizePeerInverseEdgesUnlocked(this); err != nil {
		return err
	}
	return finalizeManyToManyRelationsUnlocked(this)
}

func normalizeAllForeignKeyMapsUnlocked(reg *SchemaRegistry) error {
	for _, sch := range reg.schemas {
		for i := range sch.relations {
			rel := &sch.relations[i]
			if err := normalizeRelationForeignKeys(reg, sch, rel); err != nil {
				return errors.Wrapf(err, "schema '%s' relation '%s'", sch.Name(), rel.Edge)
			}
		}
	}
	return nil
}

func normalizeRelationForeignKeys(reg *SchemaRegistry, owner *ModelSchema, rel *ModelRelation) error {
	if rel.RelationType == RelationTypeManyToMany || rel.InversePeerSchemaName != "" {
		return nil
	}
	if len(rel.UnvalidatedFkMap) == 0 {
		backfillForeignKeysFromLegacy(rel)
		syncLegacySrcDestFromForeignKeys(rel)
		return nil
	}
	dest := reg.schemas[rel.DestSchemaName]
	if dest == nil {
		return errors.Errorf("destination schema '%s' not found in registry", rel.DestSchemaName)
	}
	pairs, err := buildForeignKeyPairsFromMap(owner, dest, rel.RelationType, rel.UnvalidatedFkMap)
	if err != nil {
		return err
	}
	rel.ForeignKeys = pairs
	rel.UnvalidatedFkMap = nil
	syncLegacySrcDestFromForeignKeys(rel)
	return nil
}

func backfillForeignKeysFromLegacy(rel *ModelRelation) {
	if len(rel.ForeignKeys) > 0 || rel.RelationType == RelationTypeManyToMany {
		return
	}
	if rel.SrcField != "" && rel.DestField != "" {
		rel.ForeignKeys = []ForeignKeyColumnPair{{
			FkColumn:         rel.SrcField,
			ReferencedColumn: rel.DestField,
		}}
	}
}

func fkMapReferencedString(val any) (string, error) {
	if val == nil {
		return "", errors.New("foreign key map: referenced column name is required")
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return fmt.Sprint(val), nil
}

func buildForeignKeyPairsFromMap(
	owner, dest *ModelSchema, relType RelationType, m DynamicFields,
) ([]ForeignKeyColumnPair, error) {
	switch relType {
	case RelationTypeManyToOne, RelationTypeOneToOne:
		return orderFkOwnerLocalToReferenced(owner, dest, m)
	case RelationTypeOneToMany:
		return orderFkOwnerChildToParent(owner, dest, m)
	default:
		return nil, errors.Errorf("unsupported relation type for foreign key map: %s", relType)
	}
}

func orderFkOwnerLocalToReferenced(owner, dest *ModelSchema, m DynamicFields) ([]ForeignKeyColumnPair, error) {
	pks := dest.PrimaryKeys()
	if len(pks) == 0 {
		return nil, errors.Errorf("schema '%s' has no primary key for FK reference", dest.Name())
	}
	pairs := make([]ForeignKeyColumnPair, 0, len(pks))
	usedLocal := map[string]struct{}{}
	for _, pk := range pks {
		local, err := localColumnForReferencedPK(m, pk, usedLocal)
		if err != nil {
			return nil, errors.Wrapf(err, "schema '%s' -> '%s'", owner.Name(), dest.Name())
		}
		if err := assertFieldOnSchema(owner, local); err != nil {
			return nil, err
		}
		if err := assertReferencedPK(dest, pk); err != nil {
			return nil, err
		}
		if err := assertMatchingTypes(owner, local, dest, pk); err != nil {
			return nil, err
		}
		pairs = append(pairs, ForeignKeyColumnPair{FkColumn: local, ReferencedColumn: pk})
	}
	if err := assertFkMapExhausted(m, usedLocal); err != nil {
		return nil, err
	}
	return pairs, nil
}

func orderFkOwnerChildToParent(parent, child *ModelSchema, m DynamicFields) ([]ForeignKeyColumnPair, error) {
	pks := parent.PrimaryKeys()
	if len(pks) == 0 {
		return nil, errors.Errorf("schema '%s' has no primary key for one-to-many reference", parent.Name())
	}
	pairs := make([]ForeignKeyColumnPair, 0, len(pks))
	usedChildCols := map[string]struct{}{}
	for _, pk := range pks {
		childCol, err := localColumnForReferencedPK(m, pk, usedChildCols)
		if err != nil {
			return nil, errors.Wrapf(err, "schema '%s' -> '%s'", parent.Name(), child.Name())
		}
		if err := assertFieldOnSchema(child, childCol); err != nil {
			return nil, err
		}
		if err := assertReferencedPK(parent, pk); err != nil {
			return nil, err
		}
		if err := assertMatchingTypes(child, childCol, parent, pk); err != nil {
			return nil, err
		}
		pairs = append(pairs, ForeignKeyColumnPair{FkColumn: childCol, ReferencedColumn: pk})
	}
	if err := assertFkMapExhausted(m, usedChildCols); err != nil {
		return nil, err
	}
	return pairs, nil
}

func localColumnForReferencedPK(m DynamicFields, pk string, used map[string]struct{}) (string, error) {
	var local string
	found := false
	for lk, rv := range m {
		ref, err := fkMapReferencedString(rv)
		if err != nil {
			return "", err
		}
		if ref != pk {
			continue
		}
		if found && local != lk {
			return "", errors.Errorf("ambiguous mapping for referenced column '%s'", pk)
		}
		local = lk
		found = true
	}
	if !found {
		return "", errors.Errorf("missing mapping for primary key column '%s'", pk)
	}
	if _, dup := used[local]; dup {
		return "", errors.Errorf("duplicate local column '%s' in foreign key map", local)
	}
	used[local] = struct{}{}
	return local, nil
}

func assertFkMapExhausted(m DynamicFields, used map[string]struct{}) error {
	for lk := range m {
		if _, ok := used[lk]; !ok {
			return errors.Errorf("unknown local column '%s' in foreign key map", lk)
		}
	}
	return nil
}

func assertFieldOnSchema(sch *ModelSchema, fieldName string) error {
	if _, ok := sch.Field(fieldName); !ok {
		return errors.Errorf("field '%s' is not defined on schema '%s'", fieldName, sch.Name())
	}
	return nil
}

func assertReferencedPK(dest *ModelSchema, col string) error {
	if !dest.IsPrimaryKey(col) {
		return errors.Errorf("'%s' is not a primary key column on schema '%s'", col, dest.Name())
	}
	return nil
}

func assertMatchingTypes(leftSch *ModelSchema, leftCol string, rightSch *ModelSchema, rightCol string) error {
	lf, _ := leftSch.Field(leftCol)
	rf, _ := rightSch.Field(rightCol)
	if lf == nil || rf == nil {
		return errors.New("assertMatchingTypes: field lookup failed")
	}
	if lf.DataType().String() != rf.DataType().String() {
		return errors.Errorf(
			"field type mismatch '%s'.%s (%s) vs '%s'.%s (%s)",
			leftSch.Name(), leftCol, lf.DataType().String(),
			rightSch.Name(), rightCol, rf.DataType().String(),
		)
	}
	return nil
}

func finalizePeerInverseEdgesUnlocked(reg *SchemaRegistry) error {
	for _, destSch := range reg.schemas {
		for i := range destSch.relations {
			rel := &destSch.relations[i]
			if rel.InversePeerSchemaName == "" {
				continue
			}
			resolved, err := resolvePeerInverseEdge(reg, destSch, rel)
			if err != nil {
				return errors.Wrapf(err, "schema '%s' peer edge '%s'", destSch.Name(), rel.Edge)
			}
			destSch.relations[i] = resolved
			addVirtualEdgeFieldOnSchema(destSch, resolved)
		}
	}
	return nil
}

func resolvePeerInverseEdge(reg *SchemaRegistry, destSch *ModelSchema, pending *ModelRelation) (ModelRelation, error) {
	srcSch := reg.schemas[pending.InversePeerSchemaName]
	if srcSch == nil {
		return *pending, errors.Errorf("peer schema '%s' not found", pending.InversePeerSchemaName)
	}
	fwd, err := findForwardPeerRelation(srcSch, pending.InversePeerEdgeName, destSch.Name())
	if err != nil {
		return *pending, err
	}
	inv, err := invertPeerRelation(*pending, fwd, srcSch.Name())
	if err != nil {
		return *pending, err
	}
	return inv, nil
}

func findForwardPeerRelation(srcSch *ModelSchema, edgeName, destSchemaName string) (ModelRelation, error) {
	for _, rel := range srcSch.Relations() {
		if rel.Edge != edgeName || rel.DestSchemaName != destSchemaName {
			continue
		}
		if rel.RelationType == RelationTypeManyToMany || rel.InversePeerSchemaName != "" {
			continue
		}
		if !array.Contains([]RelationType{
			RelationTypeManyToOne, RelationTypeOneToOne, RelationTypeOneToMany,
		}, rel.RelationType) {
			continue
		}
		return rel, nil
	}
	return ModelRelation{}, errors.Errorf(
		"no forward relation with edge '%s' from '%s' to '%s'",
		edgeName, srcSch.Name(), destSchemaName,
	)
}

func invertPeerRelation(pending ModelRelation, fwd ModelRelation, forwardOwnerSchema string) (ModelRelation, error) {
	pairs := fwd.EffectiveForeignKeys()
	if len(pairs) == 0 {
		return ModelRelation{}, errors.New("forward relation has no foreign key columns (finalize order?)")
	}
	inv := ModelRelation{
		Edge:           pending.Edge,
		label:          pending.label,
		DestSchemaName: forwardOwnerSchema,
		ForeignKeys:    pairs,
		OnDelete:       fwd.OnDelete,
		OnUpdate:       fwd.OnUpdate,
	}
	switch fwd.RelationType {
	case RelationTypeManyToOne:
		inv.RelationType = RelationTypeOneToMany
	case RelationTypeOneToOne:
		inv.RelationType = RelationTypeOneToOne
	case RelationTypeOneToMany:
		inv.RelationType = RelationTypeManyToOne
	default:
		return ModelRelation{}, errors.Errorf("cannot invert relation type %s", fwd.RelationType)
	}
	inv.InversePeerSchemaName = ""
	inv.InversePeerEdgeName = ""
	syncLegacySrcDestFromForeignKeys(&inv)
	return inv, nil
}
