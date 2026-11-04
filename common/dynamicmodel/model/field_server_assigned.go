package model

// MetadataServerAssigned flags a field whose value the owning domain service stamps before the write
// (a document number issued from a sequence), so a client never supplies it. meta/schema reports it as
// auto-generated and a form offers no input. It is deliberately not `auto_generated`: Validate empties
// an auto-generated field on create so its default can fill it, which would erase the stamped value.
// The field stays `required_for_create` where a unique constraint needs it; the stamp satisfies that.
const MetadataServerAssigned = "server_assigned"

// IsServerAssigned reports whether the field carries "metadata": {"server_assigned": true}.
func (this *ModelField) IsServerAssigned() bool {
	value, found := this.MetadataValue(MetadataServerAssigned)
	if !found {
		return false
	}
	flag, isBool := value.(bool)
	return isBool && flag
}
