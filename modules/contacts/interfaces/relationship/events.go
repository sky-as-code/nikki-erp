package relationship

type RelationshipCreatedEvent struct {
	ID            string `json:"id"`
	TargetPartyID string `json:"target_party_id"`
	Type          string `json:"type"`
	Note          string `json:"note,omitempty"`
	CreatedBy     string `json:"created_by,omitempty"`
	EventID       string `json:"event_id"`
}

type RelationshipUpdatedEvent struct {
	ID            string `json:"id"`
	TargetPartyID string `json:"target_party_id,omitempty"`
	Type          string `json:"type,omitempty"`
	Note          string `json:"note,omitempty"`
	UpdatedBy     string `json:"updated_by,omitempty"`
	EventID       string `json:"event_id"`
}

type RelationshipDeletedEvent struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
