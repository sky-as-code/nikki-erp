package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// DeriveCustomerIdentityMode calculates the identity snapshot an order is created with, from the
// request's own authentication and from nothing a caller supplied.
//
// Only a user principal counts as an identified buyer. A service principal is authenticated too,
// but it is a workload — a kiosk, a job, a message consumer — and the machine having credentials
// says nothing about who is standing in front of it. Reading it as an identified customer would let
// every kiosk sale claim policies written for people who can be contacted afterwards, which is
// exactly the entitlement an anonymous walk-up sale must not get.
//
// Absence therefore answers `anonymous`, and the caller cannot override it.
func DeriveCustomerIdentityMode(ctx corectx.Context) models.CustomerIdentityMode {
	if ctx == nil {
		return models.CustomerIdentityModeAnonymous
	}
	principal := ctx.GetPermissions().Principal
	if principal.IsZero() || principal.Kind != corectx.PrincipalKindUser {
		return models.CustomerIdentityModeAnonymous
	}
	return models.CustomerIdentityModeAuthenticated
}
