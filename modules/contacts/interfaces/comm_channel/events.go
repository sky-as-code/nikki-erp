package comm_channel

type CommChannelCreatedEvent struct {
	ID        string `json:"id"`
	PartyID   string `json:"party_id"`
	Type      string `json:"type"`
	Value     string `json:"value,omitempty"`
	Note      string `json:"note,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	EventID   string `json:"event_id"`
}

type CommChannelUpdatedEvent struct {
	ID        string `json:"id"`
	PartyID   string `json:"party_id,omitempty"`
	Type      string `json:"type,omitempty"`
	Value     string `json:"value,omitempty"`
	Note      string `json:"note,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	EventID   string `json:"event_id"`
}

type CommChannelDeletedEvent struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
