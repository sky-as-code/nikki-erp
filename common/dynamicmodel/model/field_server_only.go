package model

// MetadataServerOnly flags a field whose value never leaves the server through another schema: a
// secret such as a key the owning module keeps for its own algorithm. The query builder does not
// expand it into a bare edge and refuses it as a leaf reached through an edge, whether selected,
// filtered or sorted, so a resource that merely references the owner cannot disclose it. Reads of
// the owner's own rows are the owning module's to guard: its root columns are not affected.
const MetadataServerOnly = "server_only"

// IsServerOnly reports whether the field carries "metadata": {"server_only": true}.
func (this *ModelField) IsServerOnly() bool {
	value, found := this.MetadataValue(MetadataServerOnly)
	if !found {
		return false
	}
	flag, isBool := value.(bool)
	return isBool && flag
}
