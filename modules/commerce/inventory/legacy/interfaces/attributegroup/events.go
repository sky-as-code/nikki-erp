package attributegroup

type AttributeGroupCreatedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	EventID   string `json:"event_id"`
}

type AttributeGroupUpdatedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventID   string `json:"event_id"`
}

type AttributeGroupDeletedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
