package user

type UserCreatedEvent struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Status      string `json:"status"`
	CreatedBy   string `json:"created_by,omitempty"`
	EventID     string `json:"event_id"`
}

type UserUpdatedEvent struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Status      string `json:"status"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	EventID     string `json:"event_id"`
}

type UserDeletedEvent struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deleted_by,omitempty"`
	EventID   string `json:"event_id"`
}
