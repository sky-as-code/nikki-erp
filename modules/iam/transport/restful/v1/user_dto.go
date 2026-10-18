package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

type GetUserContextResponse struct {
	Id           string   `json:"id"`
	AvatarUrl    *string  `json:"avatar_url"`
	DisplayName  string   `json:"display_name"`
	Email        string   `json:"email"`
	Entitlements []string `json:"entitlements"`

	// The caller's own evaluation context. The frontend mirrors the guard's candidate-expression
	// algorithm to decide what to show, and entitlements alone are not enough to run it: a bare
	// `org` grant answers only for an org the caller belongs to, and a private grant only for
	// their own record. Without these the mirror would refuse things the backend allows.
	IsOwner      bool     `json:"is_owner"`
	UserOrgIds   []string `json:"user_org_ids"`
	OrgUnitId    *string  `json:"org_unit_id"`
	OrgUnitOrgId *string  `json:"org_unit_org_id"`

	Orgs            []dmodel.DynamicFields `json:"orgs"`
	AccountSettings map[string]any         `json:"account_settings"`
	SystemSettings  map[string]any         `json:"system_settings"`
}

