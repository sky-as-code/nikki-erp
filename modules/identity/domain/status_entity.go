package domain

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type IdentityStatus = enum.Enum

func WrapIdentStatus(status *enum.Enum) *IdentityStatus {
	return (*IdentityStatus)(status)
}

func WrapIdentStatuses(statuses []enum.Enum) []IdentityStatus {
	return array.Map(statuses, func(status enum.Enum) IdentityStatus {
		return *WrapIdentStatus(&status)
	})
}

const (
	OrgStatusEnumType = "ident_org_status"

	OrgStatusActive   = "active"
	OrgStatusArchived = "archived"
)

const (
	UserStatusEnumType = "ident_user_status"

	UserStatusActive   = "active"
	UserStatusArchived = "archived"
	UserStatusLocked   = "locked"
)
