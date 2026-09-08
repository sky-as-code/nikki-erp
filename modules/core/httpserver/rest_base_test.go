package httpserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// An insufficient-permission refusal must answer 403 through every payload shape the
// handlers actually pass, because the hand-written routes and the generic engine routes
// build that payload differently. If one shape falls through to 400, the two surfaces
// disagree about what a refusal looks like.
func TestClientErrorStatus_AuthorizationIsForbidden(t *testing.T) {
	refusal := *ft.NewInsufficientPermissionsError([]string{"read:iam_role:tenant"})
	collection := ft.ClientErrors{refusal}

	tests := map[string]any{
		"ClientErrors value":   collection,
		"ClientErrors pointer": &collection,
		"single item value":    refusal,
		"single item pointer":  &refusal,
		"loose any slice":      []any{refusal},
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusForbidden, clientErrorStatus(payload))
		})
	}
}

// A caller nothing authenticated must answer 401, not 403, through every payload shape.
//
// The two are both typed `authorization`, and the difference decides what the client does: a 401
// means sign in, a 403 means you are signed in and may not do this. Answering 403 for an expired
// token leaves the client stuck; answering 401 for a plain refusal signs the user out for opening
// a page containing one action they cannot perform.
func TestClientErrorStatus_UnauthenticatedIsUnauthorizedStatus(t *testing.T) {
	unauthenticated := *ft.NewUnauthenticatedError()
	collection := ft.ClientErrors{unauthenticated}

	tests := map[string]any{
		"ClientErrors value":   collection,
		"ClientErrors pointer": &collection,
		"single item value":    unauthenticated,
		"single item pointer":  &unauthenticated,
		"loose any slice":      []any{unauthenticated},
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, clientErrorStatus(payload))
		})
	}
}

// The JWT middleware's own refusals are unauthenticated too, so they must not degrade to 403.
func TestClientErrorStatus_TokenErrorsAreUnauthenticated(t *testing.T) {
	for _, key := range []string{"err_invalid_access_token", "err_malformed_access_token"} {
		t.Run(key, func(t *testing.T) {
			item := *ft.NewAuthorizationError(ft.ErrorKey(key), "bad token")
			assert.Equal(t, http.StatusUnauthorized, clientErrorStatus(ft.ClientErrors{item}))
		})
	}
}

// Mixed payloads answer 401: an unauthenticated caller cannot be told which grants they lack,
// because nothing established who they are.
func TestClientErrorStatus_UnauthenticatedWinsOverForbidden(t *testing.T) {
	mixed := ft.ClientErrors{
		*ft.NewInsufficientPermissionsError([]string{"read:iam_role:tenant"}),
		*ft.NewUnauthenticatedError(),
	}

	assert.Equal(t, http.StatusUnauthorized, clientErrorStatus(mixed))
}

// Everything that is genuinely a bad request must keep answering 400. Upgrading a
// validation failure to 403 would tell the caller they lack permission to send a payload
// they are perfectly entitled to send — only its contents were wrong.
func TestClientErrorStatus_NonAuthorizationStaysBadRequest(t *testing.T) {
	validation := *ft.NewValidationError("email", ft.ErrorKey("err_required"), "required")
	business := *ft.NewAnonymousNotFoundError()

	tests := map[string]any{
		"validation error": ft.ClientErrors{validation},
		"business error":   ft.ClientErrors{business},
		"mixed non-auth":   ft.ClientErrors{validation, business},
		"empty collection": ft.ClientErrors{},
		"unknown shape":    "something else entirely",
		"nil payload":      nil,
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, clientErrorStatus(payload))
		})
	}
}

// A refusal mixed in with validation noise is still a refusal. The caller cannot act on
// the validation complaints until they hold the permission, so the stronger answer wins.
func TestClientErrorStatus_AuthorizationWinsWhenMixed(t *testing.T) {
	mixed := ft.ClientErrors{
		*ft.NewValidationError("email", ft.ErrorKey("err_required"), "required"),
		*ft.NewInsufficientPermissionsError([]string{"read:iam_role:domain"}),
	}

	assert.Equal(t, http.StatusForbidden, clientErrorStatus(mixed))
}

func TestClientErrorStatus_NilTypedPointersAreSafe(t *testing.T) {
	var nilCollection *ft.ClientErrors
	var nilItem *ft.ClientErrorItem

	assert.Equal(t, http.StatusBadRequest, clientErrorStatus(nilCollection))
	assert.Equal(t, http.StatusBadRequest, clientErrorStatus(nilItem))
}
