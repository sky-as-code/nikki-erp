package party

type PartyCreatedEvent struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	LegalName   string `json:"legal_name,omitempty"`
	Type        string `json:"type"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	JobPosition string `json:"job_position,omitempty"`
	Website     string `json:"website,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	EventID     string `json:"event_id"`
}

type PartyUpdatedEvent struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	LegalName   string `json:"legal_name,omitempty"`
	Type        string `json:"type,omitempty"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	JobPosition string `json:"job_position,omitempty"`
	Website     string `json:"website,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	EventID     string `json:"event_id"`
}

type PartyDeletedEvent struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
