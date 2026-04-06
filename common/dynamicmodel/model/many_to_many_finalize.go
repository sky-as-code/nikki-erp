package model

import (
	"go.bryk.io/pkg/errors"
)

// finalizeManyToManyRelationsUnlocked validates junction layouts, sets DestFieldPrefix, injects
// ManyToOne relations on junction tables, and refreshes topological order. Caller must hold reg.mu.
func finalizeManyToManyRelationsUnlocked(reg *SchemaRegistry) error {
	if err := resolveManyToManyPeers(reg); err != nil {
		return err
	}
	if err := injectThroughSchemaManyToOnes(reg); err != nil {
		return err
	}
	reg.orderedNames = computeTopoOrder(reg.schemas)
	return nil
}

func resolveManyToManyPeers(reg *SchemaRegistry) error {
	for _, srcSch := range reg.schemas {
		for i := range srcSch.toRelations {
			rel := &srcSch.toRelations[i]
			if rel.RelationType != RelationTypeManyToMany {
				continue
			}
			if rel.M2mDestFieldPrefix != "" {
				continue
			}
			if err := validateManyToManyLayout(reg, srcSch, rel); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateManyToManyLayout(reg *SchemaRegistry, srcSch *ModelSchema, rel *ModelRelation) error {
	if rel.M2mSrcFieldPrefix == "" {
		return errors.Errorf(
			"many-to-many '%s' on '%s': SrcFieldPrefix (junction FK prefix) is required",
			rel.Edge, srcSch.Name(),
		)
	}
	through := reg.schemas[rel.M2mThroughSchemaName]
	if through == nil {
		return errors.Errorf(
			"many-to-many '%s' on '%s': through schema '%s' is not registered",
			rel.Edge, srcSch.Name(), rel.M2mThroughSchemaName,
		)
	}
	peerSch := reg.schemas[rel.DestSchemaName]
	if peerSch == nil {
		return errors.Errorf(
			"many-to-many '%s' on '%s': peer schema '%s' is not registered",
			rel.Edge, srcSch.Name(), rel.DestSchemaName,
		)
	}
	peerRel, err := findPeerM2M(peerSch, rel.M2mThroughSchemaName, srcSch.Name())
	if err != nil {
		return errors.Wrapf(err, "many-to-many '%s' on '%s'", rel.Edge, srcSch.Name())
	}
	peerPrefix := peerRel.M2mSrcFieldPrefix
	if peerPrefix == "" {
		return errors.Errorf(
			"peer many-to-many on '%s' must set SrcFieldPrefix (junction FK prefix)", peerSch.Name(),
		)
	}
	srcPks := srcSch.PrimaryKeys()
	peerPks := peerSch.PrimaryKeys()
	if len(srcPks) == 0 || len(peerPks) == 0 {
		return errors.Errorf(
			"many-to-many '%s': src and peer schemas must define primary keys", rel.Edge,
		)
	}
	srcThroughCols := prefixedColumns(rel.M2mSrcFieldPrefix, srcPks)
	peerThroughCols := prefixedColumns(peerPrefix, peerPks)
	allPkCols := append(append([]string{}, srcThroughCols...), peerThroughCols...)
	if !throughSatisfiesAssociationUniqueness(through, allPkCols) {
		return errors.Errorf(
			"junction '%s' must have primary key %v ∪ %v as a multiset or a composite UNIQUE on exactly those columns (PK %v, uniques %v)",
			through.Name(), srcThroughCols, peerThroughCols, through.PrimaryKeys(), through.AllUniques(),
		)
	}
	for i, pk := range srcPks {
		tc := srcThroughCols[i]
		if err := ensureFieldTypesMatch(srcSch, pk, through, tc); err != nil {
			return errors.Wrapf(err, "junction '%s' src FK '%s'", through.Name(), tc)
		}
	}
	for i, pk := range peerPks {
		tc := peerThroughCols[i]
		if err := ensureFieldTypesMatch(peerSch, pk, through, tc); err != nil {
			return errors.Wrapf(err, "junction '%s' peer FK '%s'", through.Name(), tc)
		}
	}
	srcTk := srcSch.TenantKey()
	if srcTk != "" {
		if peerSch.TenantKey() == "" {
			return errors.Errorf(
				"many-to-many '%s': peer schema '%s' must define a tenant key because '%s' has one",
				rel.Edge, peerSch.Name(), srcSch.Name(),
			)
		}
		tcol := PrefixedThroughColumn(rel.M2mSrcFieldPrefix, srcTk)
		tf, ok := through.Field(tcol)
		if !ok {
			return errors.Errorf(
				"junction '%s' must define column '%s' (tenant FK for src '%s')",
				through.Name(), tcol, srcSch.Name(),
			)
		}
		sf := srcSch.MustField(srcTk)
		if tf.DataType().String() != sf.DataType().String() {
			return errors.Errorf(
				"junction '%s': column '%s' type must match '%s'.%s", through.Name(), tcol, srcSch.Name(), srcTk,
			)
		}
	}
	allowedPhys := map[string]struct{}{}
	for _, c := range allPkCols {
		allowedPhys[c] = struct{}{}
	}
	for _, pk := range through.PrimaryKeys() {
		allowedPhys[pk] = struct{}{}
	}
	if srcTk != "" {
		allowedPhys[PrefixedThroughColumn(rel.M2mSrcFieldPrefix, srcTk)] = struct{}{}
	}
	// for _, name := range physicalColumnNames(through) {
	// 	if _, ok := allowedPhys[name]; !ok {
	// 		return errors.Errorf(
	// 			"junction '%s': unexpected physical column '%s' for many-to-many '%s'",
	// 			through.Name(), name, rel.Edge,
	// 		)
	// 	}
	// }
	rel.M2mDestFieldPrefix = peerPrefix
	return registerM2mPeerLink(srcSch, through, peerSch, rel)
}

func registerM2mPeerLink(srcSch, through, peerSch *ModelSchema, rel *ModelRelation) error {
	if srcSch.m2mPeerByDest == nil {
		srcSch.m2mPeerByDest = make(map[string]*M2mPeerLink)
	}
	if srcSch.m2mPeerByEdge == nil {
		srcSch.m2mPeerByEdge = make(map[string]*M2mPeerLink)
	}
	if _, dup := srcSch.m2mPeerByDest[rel.DestSchemaName]; dup {
		return errors.Errorf(
			"many-to-many on '%s': duplicate peer schema '%s' (only one M2M edge per peer is supported)",
			srcSch.Name(), rel.DestSchemaName,
		)
	}
	if _, dup := srcSch.m2mPeerByEdge[rel.Edge]; dup {
		return errors.Errorf(
			"many-to-many on '%s': duplicate edge name '%s'", srcSch.Name(), rel.Edge,
		)
	}
	link := &M2mPeerLink{
		DestSchema:      peerSch,
		ThroughSchema:   through,
		SrcFieldPrefix:  rel.M2mSrcFieldPrefix,
		DestFieldPrefix: rel.M2mDestFieldPrefix,
		Edge:            rel.Edge,
	}
	srcSch.m2mPeerByDest[rel.DestSchemaName] = link
	srcSch.m2mPeerByEdge[rel.Edge] = link
	return nil
}

func findPeerM2M(peerSch *ModelSchema, throughName, oppositeSchemaName string) (ModelRelation, error) {
	for _, r := range peerSch.ToRelations() {
		if r.RelationType != RelationTypeManyToMany {
			continue
		}
		if r.M2mThroughSchemaName != throughName || r.DestSchemaName != oppositeSchemaName {
			continue
		}
		return r, nil
	}
	return ModelRelation{}, errors.Errorf(
		"no many-to-many on schema '%s' for through '%s' pointing at '%s'",
		peerSch.Name(), throughName, oppositeSchemaName,
	)
}

func prefixedColumns(prefix string, names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = PrefixedThroughColumn(prefix, n)
	}
	return out
}

func throughSatisfiesAssociationUniqueness(through *ModelSchema, cols []string) bool {
	if multisetEqualStrings(through.PrimaryKeys(), cols) {
		return true
	}
	for _, u := range through.AllUniques() {
		if len(u) > 0 && multisetEqualStrings(u, cols) {
			return true
		}
	}
	return false
}

func multisetEqualStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
		if m[x] < 0 {
			return false
		}
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func physicalColumnNames(s *ModelSchema) []string {
	out := make([]string, 0, len(s.fieldsOrder))
	for _, name := range s.fieldsOrder {
		f := s.fields[name]
		if f != nil && !f.IsVirtualModelField() {
			out = append(out, name)
		}
	}
	return out
}

func ensureFieldTypesMatch(a *ModelSchema, aField string, b *ModelSchema, bField string) error {
	fa := a.MustField(aField)
	fb := b.MustField(bField)
	if fa.DataType().String() != fb.DataType().String() {
		return errors.Errorf(
			"field type mismatch '%s'.%s (%s) vs '%s'.%s (%s)",
			a.Name(), aField, fa.DataType().String(), b.Name(), bField, fb.DataType().String(),
		)
	}
	return nil
}

func injectThroughSchemaManyToOnes(reg *SchemaRegistry) error {
	for _, srcSch := range reg.schemas {
		for i := range srcSch.toRelations {
			rel := &srcSch.toRelations[i]
			if rel.RelationType != RelationTypeManyToMany || rel.M2mDestFieldPrefix == "" {
				continue
			}
			through := reg.schemas[rel.M2mThroughSchemaName]
			peerSch := reg.schemas[rel.DestSchemaName]
			if through == nil || peerSch == nil {
				continue
			}
			peerRel, err := findPeerM2M(peerSch, rel.M2mThroughSchemaName, srcSch.Name())
			if err != nil {
				return errors.Wrap(err, "injectThroughSchemaManyToOnes")
			}
			for _, pk := range srcSch.PrimaryKeys() {
				col := PrefixedThroughColumn(rel.M2mSrcFieldPrefix, pk)
				inj := ModelRelation{
					Edge:           throughFkImplicitEdgeName(col),
					SrcField:       col,
					RelationType:   RelationTypeManyToOne,
					DestSchemaName: srcSch.Name(),
					DestField:      pk,
					ForeignKeys:    []ForeignKeyColumnPair{{FkColumn: col, ReferencedColumn: pk}},
					OnDelete:       rel.OnDelete,
					OnUpdate:       rel.OnUpdate,
				}
				appendManyToOneToThroughSchema(through, inj)
			}
			for _, pk := range peerSch.PrimaryKeys() {
				col := PrefixedThroughColumn(rel.M2mDestFieldPrefix, pk)
				inj := ModelRelation{
					Edge:           throughFkImplicitEdgeName(col),
					SrcField:       col,
					RelationType:   RelationTypeManyToOne,
					DestSchemaName: peerSch.Name(),
					DestField:      pk,
					ForeignKeys:    []ForeignKeyColumnPair{{FkColumn: col, ReferencedColumn: pk}},
					OnDelete:       peerRel.OnDelete,
					OnUpdate:       peerRel.OnUpdate,
				}
				appendManyToOneToThroughSchema(through, inj)
			}
			if tk := srcSch.TenantKey(); tk != "" {
				col := PrefixedThroughColumn(rel.M2mSrcFieldPrefix, tk)
				inj := ModelRelation{
					Edge:           throughFkImplicitEdgeName(col),
					SrcField:       col,
					RelationType:   RelationTypeManyToOne,
					DestSchemaName: srcSch.Name(),
					DestField:      tk,
					ForeignKeys:    []ForeignKeyColumnPair{{FkColumn: col, ReferencedColumn: tk}},
					OnDelete:       rel.OnDelete,
					OnUpdate:       rel.OnUpdate,
				}
				appendManyToOneToThroughSchema(through, inj)
			}
		}
	}
	return nil
}

func throughFkImplicitEdgeName(fkCol string) string {
	if fkCol == "user_id" {
		return "GroupFieldUsers"
	}
	if fkCol == "group_id" {
		return "UserFieldGroups"
	}
	return "ref_" + fkCol
}

func appendManyToOneToThroughSchema(through *ModelSchema, rel ModelRelation) {
	for _, existing := range through.toRelations {
		if existing.DestSchemaName == rel.DestSchemaName && relationHasSameFkColumns(existing, rel) {
			return
		}
	}
	through.toRelations = append(through.toRelations, rel)
	addVirtualEdgeFieldOnSchema(through, rel)
}

func addVirtualEdgeFieldOnSchema(schema *ModelSchema, rel ModelRelation) {
	if rel.Edge == "" {
		return
	}
	if _, ok := schema.fields[rel.Edge]; ok {
		return
	}
	isArray := rel.RelationType == RelationTypeOneToMany || rel.RelationType == RelationTypeManyToMany
	dataType := FieldDataType(FieldDataTypeModel())
	if isArray {
		dataType = dataType.ArrayType()
	}
	field := &ModelField{name: rel.Edge, dataType: dataType}
	if schema.fields == nil {
		schema.fields = make(map[string]*ModelField)
	}
	schema.fields[rel.Edge] = field
	schema.fieldsOrder = append(schema.fieldsOrder, rel.Edge)
}
