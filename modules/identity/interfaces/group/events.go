package group

type GroupCreatedEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	EventID     string `json:"event_id"`
}

type GroupUpdatedEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Etag        string `json:"etag,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	EventID     string `json:"event_id"`
}

type GroupDeletedEvent struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
