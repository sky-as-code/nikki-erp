package attributevalue

type AttributeValueCreatedEvent struct {
	Id          string `json:"id"`
	AttributeId string `json:"attribute_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	EventID     string `json:"event_id"`
}

type AttributeValueUpdatedEvent struct {
	Id          string `json:"id"`
	AttributeId string `json:"attribute_id,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	EventID     string `json:"event_id"`
}

type AttributeValueDeletedEvent struct {
	Id          string `json:"id"`
	AttributeId string `json:"attribute_id,omitempty"`
	DeletedBy   string `json:"deleted_by,omitempty"`
	EventID     string `json:"event_id"`
}
