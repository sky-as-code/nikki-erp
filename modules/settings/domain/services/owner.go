package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// tenantIdField is the tenant column coremart's base model injects. Settings never declares it and
// never filters on it — the multi-tenant repository does that for every read and write. It is named
// here only because the fan-out (which writes many owners' rows in one statement) has to scope
// itself to the acting tenant explicitly.
const tenantIdField = "tenant_id"

// ownerIdFor resolves which owner a request is acting as.
//
// A user and an organization are addressed by the request's own identity, so neither can be spoofed
// by a caller naming someone else's id. The tenant is not an entity in this module's world at all:
// it arrives as the domain constraint the multi-tenant layer already applies to every query, which
// is why a tenant-owned row is keyed by it.
func ownerIdFor(ctx corectx.Context, ownerType string) (string, error) {
	switch ownerType {
	case c.OwnerTypeUser:
		userId := ctx.GetPermissions().UserId
		if userId == "" {
			return "", errors.New("ownerIdFor: the request carries no user id")
		}
		return string(userId), nil

	case c.OwnerTypeOrg:
		orgId := actingOrgId(ctx)
		if orgId == "" {
			return "", errors.New("ownerIdFor: the request carries no organization id")
		}
		return orgId, nil

	case c.OwnerTypeTenant:
		tenantId := actingTenantId(ctx)
		if tenantId == "" {
			// The nikkierp binary has no tenant key, so there is no tenant to own a row.
			return "", errors.New("ownerIdFor: the request carries no tenant id")
		}
		return tenantId, nil
	}
	return "", errors.Errorf("ownerIdFor: unknown owner type '%s'", ownerType)
}

// actingOrgId returns the single organization the request acts for.
//
// A user may belong to several organizations, and settings deliberately refuses to guess among
// them: writing one org's configuration because it happened to sort first would be worse than
// failing. The org unit's organization is used when the request has one, since that is an explicit
// choice rather than an inference.
func actingOrgId(ctx corectx.Context) string {
	permissions := ctx.GetPermissions()
	if permissions.OrgUnitOrgId != nil && *permissions.OrgUnitOrgId != "" {
		return string(*permissions.OrgUnitOrgId)
	}
	if permissions.UserOrgIds != nil && permissions.UserOrgIds.Length() == 1 {
		return string(permissions.UserOrgIds.ToSlice()[0])
	}
	return ""
}

// actingTenantId reads the tenant from the request's domain constraints, which is where the
// multi-tenant layer puts it. It is empty in the nikkierp binary, which has no tenant key.
func actingTenantId(ctx corectx.Context) string {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil {
		return ""
	}
	raw, ok := constraints[tenantIdField]
	if !ok || raw == nil {
		return ""
	}
	return toIdString(raw)
}

// toIdString accepts the several shapes an id takes as it crosses the dynamic-model boundary.
func toIdString(raw any) string {
	switch typed := raw.(type) {
	case string:
		return typed
	case *string:
		if typed == nil {
			return ""
		}
		return *typed
	case dmodel.DynamicFields:
		return ""
	}
	if stringer, ok := raw.(interface{ String() string }); ok {
		return stringer.String()
	}
	return ""
}
