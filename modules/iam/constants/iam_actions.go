package constants

const (
	ActionView   = "View"
	ActionCreate = "Create"
	ActionUpdate = "Update"
	ActionDelete = "Delete"

	// ActionManageCredentials covers issuing a temporary password and reading it
	// back. Separate from Update because it is an account-takeover capability.
	ActionManageCredentials = "manage_credentials"
)
