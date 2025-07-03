package organization

type DeleteOrganizationEvent struct {
	Id       string `json:"id"`
	DeleteBy string `json:"delete_by,omitempty"`
	EventID  string `json:"event_id"`
}
