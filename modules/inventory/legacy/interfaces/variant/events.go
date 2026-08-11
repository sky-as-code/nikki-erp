package variant

type VariantCreatedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	EventID   string `json:"event_id"`
}

type VariantUpdatedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventID   string `json:"event_id"`
}

type VariantDeletedEvent struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id,omitempty"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
