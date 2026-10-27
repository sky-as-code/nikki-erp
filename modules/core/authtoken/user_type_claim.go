package authtoken

import (
	"github.com/golang-jwt/jwt/v5"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// JwtClaimUserType carries the IAM user type (`user` | `service`) of the token's subject in the
// access and refresh tokens. Its values are the principal kinds they map to.
const JwtClaimUserType = "usrtype"

// UserTypeClaims is the custom-claim set CreateJwt needs to carry userType.
func UserTypeClaims(userType string) map[string]any {
	return map[string]any{JwtClaimUserType: userType}
}

// PrincipalKindFromClaims maps the user type claim to the principal kind of the request. A token
// without the claim predates it and was issued to a person, so it answers user, which keeps every
// token issued before the claim existed working exactly as it did.
func PrincipalKindFromClaims(claims jwt.Claims) corectx.PrincipalKind {
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return corectx.PrincipalKindUser
	}
	if userType, _ := mapClaims[JwtClaimUserType].(string); userType == string(corectx.PrincipalKindService) {
		return corectx.PrincipalKindService
	}
	return corectx.PrincipalKindUser
}
