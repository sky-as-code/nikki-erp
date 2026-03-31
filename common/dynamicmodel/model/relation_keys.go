package model

// EffectiveForeignKeys returns the resolved FK column pairs for this relation.
func (this ModelRelation) EffectiveForeignKeys() []ForeignKeyColumnPair {
	if len(this.ForeignKeys) > 0 {
		out := make([]ForeignKeyColumnPair, len(this.ForeignKeys))
		copy(out, this.ForeignKeys)
		return out
	}
	if this.SrcField != "" && this.DestField != "" {
		return []ForeignKeyColumnPair{{
			FkColumn:         this.SrcField,
			ReferencedColumn: this.DestField,
		}}
	}
	return nil
}

func syncLegacySrcDestFromForeignKeys(rel *ModelRelation) {
	pairs := rel.EffectiveForeignKeys()
	if len(pairs) == 0 {
		return
	}
	rel.SrcField = pairs[0].FkColumn
	rel.DestField = pairs[0].ReferencedColumn
}

// RelationsShareForeignKeyColumns reports whether two relations use the same FK column pairs.
func RelationsShareForeignKeyColumns(a, b ModelRelation) bool {
	return relationHasSameFkColumns(a, b)
}

func relationHasSameFkColumns(a, b ModelRelation) bool {
	pa, pb := a.EffectiveForeignKeys(), b.EffectiveForeignKeys()
	if len(pa) != len(pb) {
		return false
	}
	for i := range pa {
		if pa[i].FkColumn != pb[i].FkColumn || pa[i].ReferencedColumn != pb[i].ReferencedColumn {
			return false
		}
	}
	return true
}
