package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every permission code the application services demand must be seeded as an iam_actions row, or
// the action is unreachable in production: the request reaches AssertAction, no grant matches, and
// the caller is refused a power an administrator believes they assigned.
//
// This is the ratchet that keeps R1 of the composable migration honest. The composable engine's
// RouteDefinition carries no permission field, so a code is a plain string an application method
// passes to AssertAction; nothing but this test ties it back to the seed. It replaces the
// equivalent coverage in dynamicengines/lifecycle_actions_test.go, which asserted the same thing
// against the legacy action definitions and is deleted with them.

// customPermissions are the codes declared in permissions.go that are not one of the engine's
// built-in verbs. Listed rather than reflected because a constant is a decision: adding one here
// is how the author states that a new power needs its own grant.
var customPermissions = []string{
	PermissionSuspend,
	PermissionActivate,
	PermissionConfirm,
	PermissionCancel,
	PermissionApplyVoucher,
	PermissionManualDiscount,
	PermissionAssignParties,
	PermissionAssignSoldTo,
	PermissionAssignBillTo,
	PermissionAssignPayer,
	PermissionSplitBill,
	PermissionMergeBill,
	PermissionPayBill,
	PermissionSettleBill,
	PermissionProcessReturn,
	PermissionConvertQuotation,
	PermissionMarkBillingReady,
	PermissionRevertBillingToDraft,
	PermissionCancelBillingInstruction,
}

func TestCustomPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t)

	for _, code := range customPermissions {
		if code == "" {
			t.Error("a permission constant is empty; an empty code makes AssertAction skip the check")
			continue
		}
		if !seeded[code] {
			t.Errorf("the %q permission is demanded by an application service but the Sales IAM "+
				"migrations do not seed it; the action would be refused for every caller", code)
		}
	}
}

// seededActionCodes parses the SQL rather than querying a database, so the test needs no database
// and fails in CI the moment a permission is added without its seed.
func seededActionCodes(t *testing.T) map[string]bool {
	t.Helper()

	// Every Sales IAM migration, not just the first: reading only 1007002 would let an action seeded
	// later pass this check while being unreachable in production.
	//
	// Matched by pattern rather than listed by name, because the incremental per-submodule migrations
	// get folded back into 1007002_sales_iam.sql. A hardcoded list breaks on the consolidation and,
	// worse, silently stops covering a file that is renamed rather than deleted.
	pattern := filepath.Join("..", "..", "..", "scripts", "migrations", "*_sales*_iam.sql")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("listing the Sales IAM migrations: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no Sales IAM migration matched %s; the migrations moved or were renamed", pattern)
	}

	codes := map[string]bool{}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("the Sales IAM migration %s must be readable from the test: %v", path, err)
		}
		collectActionCodes(string(content), codes)
	}

	if len(codes) == 0 {
		t.Fatal("no action codes parsed out of the IAM migrations; the parser or the files changed")
	}
	return codes
}

func collectActionCodes(content string, codes map[string]bool) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "('01M3SALES") {
			continue
		}
		// ('<id>', '<name>', '<code>', ...) - the code is the third quoted field.
		parts := strings.Split(trimmed, "'")
		if len(parts) < 6 {
			continue
		}
		codes[parts[5]] = true
	}
}
