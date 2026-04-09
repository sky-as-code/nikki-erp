package middlewares

type contextKey struct {
	name string
}

var CtxKeyJwtClaims = &contextKey{"JwtClaims"}
