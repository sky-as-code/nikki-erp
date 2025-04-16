package user

type User struct {
	ID                  string
	Username            string
	Email               string
	DisplayName         string
	PasswordHash        string
	AvatarUrl           string
	Status              string
	LastLoginAt         *string
	CreatedAt           string
	UpdatedAt           string
	CreatedBy           string
	UpdatedBy           string
	DeletedAt           *string
	DeletedBy           string
	MustChangePassword  bool
	PasswordChangedAt   *string
	FailedLoginAttempts int
	LockedUntil         *string
}
