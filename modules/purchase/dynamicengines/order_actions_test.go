package dynamicengines

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// A permission code the engine demands but the seed never created denies every request for that
// action, with nothing in the response pointing at the seed. This reads the migration and checks
// the Go constants and the SQL agree, which is the only comparison possible without a database.
func TestOrderActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t, "purchase_order")

	for _, permission := range []string{
		PermissionConfirm,
		PermissionApprove,
		PermissionCancel,
		PermissionSend,
		PermissionLock,
		PermissionUnlock,
		PermissionAcknowledge,
		PermissionMerge,
	} {
		assert.Contains(t, seeded, permission,
			"the engine demands the %q permission, which 1006002_purchase_iam.sql does not seed", permission)
	}

	// Duplicate reuses `create`, so the seed must not have grown a row of its own: an unused action
	// row is a permission an administrator can grant that does nothing.
	assert.NotContains(t, seeded, "duplicate")
	assert.NotContains(t, seeded, "print")
}

// The reverse direction: every code the seed creates must be one the engine demands, or it grants
// nothing.
func TestSeededOrderActionsAreDemanded(t *testing.T) {
	demanded := map[string]bool{
		drif.PermissionCreate: true,
		drif.PermissionUpdate: true,
		drif.PermissionDelete: true,
		drif.PermissionRead:   true,
		PermissionConfirm:     true,
		PermissionApprove:     true,
		PermissionCancel:      true,
		PermissionSend:        true,
		PermissionLock:        true,
		PermissionUnlock:      true,
		PermissionAcknowledge: true,
		// Merge is seeded but not yet reached; the constant exists so the two cannot drift apart.
		PermissionMerge: true,
	}

	for _, code := range seededActionCodes(t, "purchase_order") {
		assert.True(t, demanded[code],
			"1006002_purchase_iam.sql seeds the %q action, which no engine action demands", code)
	}
}

// seededActionCodes parses the IAM migration SQL rather than connecting to a database, so a
// mismatch is caught at `go test` time instead of as an unexplained permission denied.
func seededActionCodes(t *testing.T, resourceCode string) []string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "scripts", "migrations", "1006002_purchase_iam.sql")
	content, err := os.ReadFile(path)
	require.NoError(t, err, "the purchase IAM migration must be readable from the test")

	sql := string(content)

	// The resource's id, then every action row pointing at it.
	resourcePattern := regexp.MustCompile(
		`\('([0-9A-Z]{26})', '[^']*', '` + regexp.QuoteMeta(resourceCode) + `',`)
	resourceMatch := resourcePattern.FindStringSubmatch(sql)
	require.NotNil(t, resourceMatch, "no %q resource row in the migration", resourceCode)
	resourceId := resourceMatch[1]

	actionPattern := regexp.MustCompile(
		`\('[0-9A-Z]{26}', '[^']*', '([a-z_]+)', (?:'[^']*'|NULL), '` + resourceId + `'`)

	codes := []string{}
	for _, match := range actionPattern.FindAllStringSubmatch(sql, -1) {
		codes = append(codes, match[1])
	}
	require.NotEmpty(t, codes, "no action rows for %q in the migration", resourceCode)
	return codes
}

// The same two-directional check for the agreement, whose action set differs from the order's.
func TestAgreementActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t, "purchase_agreement")

	for _, permission := range []string{
		PermissionConfirm,
		PermissionClose,
		PermissionCancel,
		drif.PermissionSetArchived,
	} {
		assert.Contains(t, seeded, permission,
			"the engine demands the %q permission, which 1006002_purchase_iam.sql does not seed",
			permission)
	}

	// create_rfq carries the order's create permission, so the seed must not have grown a row for it.
	assert.NotContains(t, seeded, "create_rfq")
	// Archive and restore reuse set_archived, for the reason the seed's header gives.
	assert.NotContains(t, seeded, "archive")
	assert.NotContains(t, seeded, "restore")
}

func TestSeededAgreementActionsAreDemanded(t *testing.T) {
	demanded := map[string]bool{
		drif.PermissionCreate:      true,
		drif.PermissionUpdate:      true,
		drif.PermissionDelete:      true,
		drif.PermissionRead:        true,
		drif.PermissionSetArchived: true,
		PermissionConfirm:          true,
		PermissionClose:            true,
		PermissionCancel:           true,
	}

	for _, code := range seededActionCodes(t, "purchase_agreement") {
		assert.True(t, demanded[code],
			"1006002_purchase_iam.sql seeds the %q action, which no engine action demands", code)
	}
}
