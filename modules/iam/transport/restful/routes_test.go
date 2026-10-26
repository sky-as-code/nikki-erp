package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	for schemaName, engineName := range map[string]string{
		models.UserSchemaName:               v1.UserEngineName,
		models.OrganizationSchemaName:       v1.OrgEngineName,
		models.OrganizationalUnitSchemaName: v1.OrgUnitEngineName,
		models.GroupSchemaName:              v1.GroupEngineName,
		models.EntitlementSchemaName:        v1.EntitlementEngineName,
		models.RoleSchemaName:               v1.RoleEngineName,
		models.ResourceSchemaName:           v1.ResourceEngineName,
		models.ActionSchemaName:             v1.ActionEngineName,
		models.RoleRequestSchemaName:        v1.GrantRequestEngineName,
	} {
		assert.Equal(t, composable.EngineDependencyName(schemaName), engineName, schemaName)
	}
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// registeredRoutes registers one resource's route table on a throwaway echo instance and returns
// "METHOD path" in registration order.
func registeredRoutes(t *testing.T, engine composable.RestEngine) []string {
	t.Helper()
	echoApp := echo.New()
	require.NoError(t, engine.RegisterRoutes(echoApp.Group("/v1/iam"), passThrough))

	routes := echoApp.Router().Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}

func TestIamCustomRoutes(t *testing.T) {
	for name, testCase := range map[string]struct {
		engine   composable.RestEngine
		expected []string
	}{
		"iam_user": {
			engine: userRestEngine(&v1.UserCrudRest{}),
			expected: []string{
				"POST /v1/iam/iam_user/:id/send_invitation",
				"POST /v1/iam/iam_user/:id/manage_roles",
				"GET /v1/iam/iam_user/:id/roles",
			},
		},
		"iam_org": {
			engine: orgRestEngine(&v1.OrgCrudRest{}),
			expected: []string{
				"POST /v1/iam/iam_org/:id/manage_users",
				"POST /v1/iam/iam_org/:id/archived",
				"DELETE /v1/iam/iam_org/:id",
			},
		},
		"iam_orgunit": {
			engine:   orgUnitRestEngine(&v1.OrgUnitCrudRest{}),
			expected: []string{"POST /v1/iam/iam_orgunit/:id/manage_users"},
		},
		"iam_group": {
			engine: groupRestEngine(&v1.GroupCrudRest{}),
			expected: []string{
				"POST /v1/iam/iam_group/:id/manage_users",
				"POST /v1/iam/iam_group/:id/manage_roles",
				"GET /v1/iam/iam_group/:id/roles",
			},
		},
		"iam_role": {
			engine: roleRestEngine(&v1.RoleCrudRest{}),
			expected: []string{
				"POST /v1/iam/iam_role/:id/manage_users",
				"POST /v1/iam/iam_role/:id/manage_groups",
				"GET /v1/iam/iam_role/describe",
			},
		},
		"iam_entitlement": {
			engine:   entitlementRestEngine(&v1.EntitlementCrudRest{}),
			expected: []string{"PATCH /v1/iam/iam_entitlement/:id", "POST /v1/iam/iam_entitlement/:id/archived"},
		},
		"iam_resource": {
			engine:   resourceRestEngine(&v1.ResourceCrudRest{}),
			expected: []string{"GET /v1/iam/iam_resource", "GET /v1/iam/iam_resource/meta/schema"},
		},
		"iam_action": {
			engine:   actionRestEngine(&v1.ActionCrudRest{}),
			expected: []string{"GET /v1/iam/iam_action", "PATCH /v1/iam/iam_action/:id"},
		},
		"iam_grant_request": {
			engine:   grantRequestRestEngine(&v1.GrantRequestCrudRest{}),
			expected: []string{"POST /v1/iam/iam_grant_request", "DELETE /v1/iam/iam_grant_request/:id"},
		},
	} {
		routes := registeredRoutes(t, testCase.engine)
		for _, route := range testCase.expected {
			assert.Contains(t, routes, route, name)
		}
	}
}

// The avatar is written as avatar_file_path through the plain update; the per-field file routes
// are gone with filefield.
func TestUserServesNoFileRoutes(t *testing.T) {
	for _, route := range registeredRoutes(t, userRestEngine(&v1.UserCrudRest{})) {
		assert.NotContains(t, route, "/files/")
	}
}
