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

// The engine asserts a permission by code, and the IAM migration seeds the codes that exist. A code
// the engine demands but the seed never created denies EVERY request for that action, and nothing
// in the response points at the seed as the cause — the caller is simply told they may not.
//
// This test reads the migration and checks the two agree. It is the one place the Go constants and
// the SQL can be compared without a database.
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
			"the engine demands the %q permission, which 0007002_purchase_iam.sql does not seed", permission)
	}

	// Duplicate reuses `create` rather than carrying a permission of its own, so the seed must NOT
	// have grown one — an unused action row is a permission an administrator can grant that does
	// nothing, which is worse than no row at all.
	assert.NotContains(t, seeded, "duplicate")
	assert.NotContains(t, seeded, "print")
}

// The reverse direction: every lifecycle code the seed creates must be one the engine actually
// demands, or it is a permission that grants nothing.
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
		// Merge is seeded and reached in [PUR-020]; the constant exists so that the two cannot
		// drift apart in the meantime.
		PermissionMerge: true,
	}

	for _, code := range seededActionCodes(t, "purchase_order") {
		assert.True(t, demanded[code],
			"0007002_purchase_iam.sql seeds the %q action, which no engine action demands", code)
	}
}

// seededActionCodes reads the action codes the IAM migration creates for one resource.
//
// It parses the SQL rather than connecting to a database, because the point is to catch the
// mismatch at `go test` time — by the time a database is involved, the mismatch shows up as a
// permission denied with no explanation.
func seededActionCodes(t *testing.T, resourceCode string) []string {
	t.Helper()

	path := filepath.Join("..", "..", "..", "scripts", "migrations", "0007002_purchase_iam.sql")
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

// The same two-directional check for the agreement. Its action set is different from the order's —
// it has set_archived and close, and no send or lock — so a shared test would prove neither.
func TestAgreementActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t, "purchase_agreement")

	for _, permission := range []string{
		PermissionConfirm,
		PermissionClose,
		PermissionCancel,
		drif.PermissionSetArchived,
	} {
		assert.Contains(t, seeded, permission,
			"the engine demands the %q permission, which 0007002_purchase_iam.sql does not seed",
			permission)
	}

	// create_rfq carries the order's create permission rather than one of its own, so the seed must
	// not have grown a row for it: an action row nobody demands is a permission that grants nothing.
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
			"0007002_purchase_iam.sql seeds the %q action, which no engine action demands", code)
	}
}
