package unit

type UnitCreatedEvent struct {
	Id        string `json:"id"`
	CreatedBy string `json:"created_by,omitempty"`
	EventID   string `json:"event_id"`
}

type UnitUpdatedEvent struct {
	Id        string `json:"id"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventID   string `json:"event_id"`
}

type UnitDeletedEvent struct {
	Id        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
