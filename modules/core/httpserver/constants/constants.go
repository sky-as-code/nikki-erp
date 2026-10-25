package constants

type contextKey struct {
	name string
}

var CtxKeyJwtClaims = &contextKey{"JwtClaims"}
var CtxKeyUserIsAuthorized = &contextKey{"UserIsAuthorized"}
