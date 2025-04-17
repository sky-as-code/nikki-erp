package user

import (
	"go.bryk.io/pkg/errors"

	ent "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/user"
)

type Status ent.Status

const DefaultStatus = ent.DefaultStatus

const (
	StatusActive    = Status(ent.StatusActive)
	StatusInactive  = Status(ent.StatusInactive)
	StatusSuspended = Status(ent.StatusSuspended)
	StatusPending   = Status(ent.StatusPending)
)

func (this Status) Validate() error {
	switch this {
	case StatusActive, StatusInactive, StatusSuspended, StatusPending:
		return nil
	default:
		return errors.Errorf("invalid status value: %s", this)
	}
}

type User struct {
	Id string

	AvatarUrl *string
	CreatedAt string
	CreatedBy string
	// DeletedAt           *string
	// DeletedBy           *string
	DisplayName         string
	Email               string
	Etag                string
	FailedLoginAttempts int
	LastLoginAt         *string
	LockedUntil         *string
	MustChangePassword  bool
	PasswordChangedAt   string
	PasswordHash        string
	Status              string
	UpdatedAt           string
	UpdatedBy           string
	Username            string
}
