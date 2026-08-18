package requestguard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func idOf(raw string) *model.Id {
	id := model.Id(raw)
	return &id
}

func TestBuildExpression(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		resource string
		scope    ResourceScope
		scopeId  *model.Id
		expected string
	}{
		{"exact domain", "create", "iam_user", ResourceScopeDomain, nil, "create:iam_user:domain"},
		{"exact private", "view", "iam_user", ResourceScopePrivate, nil, "view:iam_user:private"},
		{"bare org", "create", "iam_user", ResourceScopeOrg, nil, "create:iam_user:org"},
		{"org with id", "create", "iam_user", ResourceScopeOrg, idOf("ORG1"), "create:iam_user:org/ORG1"},
		{"bare orgunit", "create", "iam_user", ResourceScopeOrgUnit, nil, "create:iam_user:orgunit"},
		{"orgunit with id", "create", "iam_user", ResourceScopeOrgUnit, idOf("OU1"), "create:iam_user:orgunit/OU1"},
		{"wildcard action", "", "iam_user", ResourceScopeDomain, nil, "*:iam_user:domain"},
		{"wildcard resource", "create", "", ResourceScopeDomain, nil, "create:*:domain"},
		{"wildcard both", "", "", ResourceScopeDomain, nil, "*:*:domain"},
		{"explicit wildcard token", Wildcard, Wildcard, ResourceScopeOrg, idOf("ORG1"), "*:*:org/ORG1"},
		{"empty scope id ignored", "create", "iam_user", ResourceScopeOrg, idOf(""), "create:iam_user:org"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, BuildExpression(test.action, test.resource, test.scope, test.scopeId))
		})
	}
}

func TestParseExpression_Valid(t *testing.T) {
	tests := []struct {
		expr     string
		action   string
		resource string
		scope    ResourceScope
		scopeId  *model.Id
	}{
		{"create:iam_user:domain", "create", "iam_user", ResourceScopeDomain, nil},
		{"view:iam_user:private", "view", "iam_user", ResourceScopePrivate, nil},
		{"create:iam_user:org", "create", "iam_user", ResourceScopeOrg, nil},
		{"create:iam_user:org/ORG1", "create", "iam_user", ResourceScopeOrg, idOf("ORG1")},
		{"create:iam_user:orgunit/OU1", "create", "iam_user", ResourceScopeOrgUnit, idOf("OU1")},
		{"*:iam_user:domain", Wildcard, "iam_user", ResourceScopeDomain, nil},
		{"create:*:domain", "create", Wildcard, ResourceScopeDomain, nil},
		{"*:*:*", Wildcard, Wildcard, ResourceScope(Wildcard), nil},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			parsed, err := ParseExpression(test.expr)
			require.NoError(t, err)
			assert.Equal(t, test.action, parsed.ActionCode)
			assert.Equal(t, test.resource, parsed.ResourceCode)
			assert.Equal(t, test.scope, parsed.Scope)
			if test.scopeId == nil {
				assert.Nil(t, parsed.ScopeId)
			} else {
				require.NotNil(t, parsed.ScopeId)
				assert.Equal(t, *test.scopeId, *parsed.ScopeId)
			}
			// Parsing then rebuilding must be the identity.
			assert.Equal(t, test.expr, parsed.String())
		})
	}
}

// Every one of these must come back as an error rather than a panic: the probe
// endpoint feeds untrusted request bodies straight into the parser, and a panic
// there is a 500 where a 400 belongs.
func TestParseExpression_Invalid(t *testing.T) {
	tests := []string{
		"",
		"create",
		"create:iam_user",
		"create:iam_user:domain:extra",
		"::",
		":iam_user:domain",
		"create::domain",
		"create:iam_user:",
		"create:iam_user:galaxy",
		"create:iam_user:domain/ORG1",
		"create:iam_user:private/ORG1",
		"create:iam_user:org/",
		"*:iam_user:*",
		"create:*:*",
		"create:iam_user:org/ORG1/EXTRA",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			parsed, err := ParseExpression(expr)
			assert.Error(t, err, "expected %q to be rejected", expr)
			assert.Nil(t, parsed)
		})
	}
}

// Found by the API security suite: a caller holding a wildcard grant could ask
// about a resource code thousands of characters long and be answered "granted",
// because a wildcard matches any code at all. Nothing legitimate is that long, and
// the segment travels on into a query parameter, so the parser bounds it.
func TestParseExpression_RejectsOverlongSegments(t *testing.T) {
	long := strings.Repeat("a", maxSegmentLength+1)

	overlong := []string{
		long + ":iam_user:domain",
		"read:" + long + ":domain",
		"read:iam_user:org/" + long,
	}
	for _, expr := range overlong {
		parsed, err := ParseExpression(expr)
		assert.Error(t, err, "an overlong segment must be rejected")
		assert.Nil(t, parsed)
	}

	// At the limit it is still valid: the bound rejects the absurd, not the long.
	atLimit := "read:" + strings.Repeat("a", maxSegmentLength) + ":domain"
	_, err := ParseExpression(atLimit)
	assert.NoError(t, err)
}

func TestParsedExpression_HasWildcard(t *testing.T) {
	tests := []struct {
		expr     string
		wildcard bool
	}{
		{"create:iam_user:domain", false},
		{"create:iam_user:org/ORG1", false},
		{"*:iam_user:domain", true},
		{"create:*:domain", true},
		{"*:*:*", true},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			parsed, err := ParseExpression(test.expr)
			require.NoError(t, err)
			assert.Equal(t, test.wildcard, parsed.HasWildcard())
		})
	}
}
