package enum

type EnumCreatedEvent struct {
	Id        string `json:"id"`
	CreatedBy string `json:"created_by,omitempty"`
	EventID   string `json:"event_id"`
}

type EnumUpdatedEvent struct {
	Id        string `json:"id"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventID   string `json:"event_id"`
}

type EnumDeletedEvent struct {
	Id        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
