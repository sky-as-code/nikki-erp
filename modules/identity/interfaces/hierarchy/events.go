package hierarchy

type HierarchyCreatedEvent struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by,omitempty"`
	EventId   string `json:"event_id"`
}
type HierarchyLevelUpdatedEvent struct {
	Id        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Etag      string `json:"etag,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventId   string `json:"event_id"`
}

type HierarchyDeletedEvent struct {
	Id        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventId   string `json:"event_id"`
}
