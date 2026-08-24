package dynamicengines

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// captureEngine records the action definitions without building a real engine.
type captureEngine struct {
	drif.DynamicResourceEngine
	defs []drif.DynamicActionDefinition
}

func (this *captureEngine) DefineAction(def drif.DynamicActionDefinition) error {
	this.defs = append(this.defs, def)
	return nil
}

func channelActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesChannelActions(engine); err != nil {
		t.Fatalf("defineSalesChannelActions: %v", err)
	}
	return engine.defs
}

func pointActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesPointActions(engine); err != nil {
		t.Fatalf("defineSalesPointActions: %v", err)
	}
	return engine.defs
}

// TestActionsAreFullyDeclared pins the shape of every lifecycle action.
//
// ActionName and MainProcess are mandatory; a definition missing either is rejected at Init, which
// means a broken build boots and then fails on the first request rather than at start-up.
func TestActionsAreFullyDeclared(t *testing.T) {
	for _, def := range append(channelActions(t), pointActions(t)...) {
		if def.ActionName == "" {
			t.Error("an action is missing its name")
		}
		if def.MainProcess == nil {
			t.Errorf("action %q has no MainProcess", def.ActionName)
		}
		if def.ActionType != drif.ActionTypeGeneric {
			t.Errorf("action %q is %q, want ActionTypeGeneric: these carry a request body and "+
				"none of them is a CRUD verb", def.ActionName, def.ActionType)
		}
		if def.RestPath == "" {
			t.Errorf("action %q has no RestPath", def.ActionName)
		}
		if strings.Contains(def.RestPath, "-") {
			t.Errorf("action %q path %q contains a hyphen; the route regex rejects them and the "+
				"word separator is underscore", def.ActionName, def.RestPath)
		}
	}
}

// TestRecordActionsAreScopedToAnId guards the difference between operating on one record and
// operating on the collection.
//
// Every lifecycle action names the record it changes. Resolve is the deliberate exception: it looks
// a channel up by code, so it has no id to be given one.
func TestRecordActionsAreScopedToAnId(t *testing.T) {
	for _, def := range append(channelActions(t), pointActions(t)...) {
		if def.ActionName == ActionResolve {
			if strings.Contains(def.RestPath, ":id") {
				t.Errorf("resolve takes a code, not an id, but its path is %q", def.RestPath)
			}
			continue
		}
		if !strings.HasPrefix(def.RestPath, ":id/") {
			t.Errorf("action %q changes one record but its path %q is collection-level",
				def.ActionName, def.RestPath)
		}
	}
}

// TestActionPermissionsAreSeeded is the check that a permission code cannot drift from its seed.
//
// The engine asserts the permission string named here against the action rows in
// 1005002_sales_iam.sql. A code with no matching row denies every request, and nothing in the 403
// points at the seed as the cause — so the two are compared directly.
func TestActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t)

	for _, def := range append(channelActions(t), pointActions(t)...) {
		if def.Permission == "" {
			t.Errorf("action %q declares no permission", def.ActionName)
			continue
		}
		if !seeded[def.Permission] {
			t.Errorf("action %q demands the %q permission, which 1005002_sales_iam.sql does not "+
				"seed", def.ActionName, def.Permission)
		}
	}
}

// seededActionCodes reads the action codes out of the IAM migration.
//
// Parsing the SQL is cruder than querying a database, and deliberately so: the test then needs no
// database, and it fails in CI the moment somebody adds an action without its seed.
func seededActionCodes(t *testing.T) map[string]bool {
	t.Helper()
	path := filepath.Join("..", "..", "..", "scripts", "migrations", "1005002_sales_iam.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the Sales IAM migration must be readable from the test: %v", err)
	}

	codes := map[string]bool{}
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "('01M3SALES") {
			continue
		}
		// ('<id>', '<name>', '<code>', ...) — the code is the third quoted field.
		parts := strings.Split(trimmed, "'")
		if len(parts) < 6 {
			continue
		}
		codes[parts[5]] = true
	}
	if len(codes) == 0 {
		t.Fatal("no action codes parsed out of the IAM migration; the parser or the file changed")
	}
	return codes
}
