package dynamicengines

import (
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// These guard the invariants the composable migration depends on. They parse nothing and touch no
// database: they read the registration lists directly, so a mistake fails in milliseconds here
// rather than at boot, where the same mistake surfaces as a startup stack trace.

// engineSchemaCount is the number of schemas Sales serves through an engine. It excludes the
// junctions, which are repository-only by design.
//
// The literal is the point: a resource added or removed without updating it is a deliberate
// decision someone must confirm, not a number that drifts silently.
const engineSchemaCount = 41

// Every schema Sales serves appears exactly once. A schema served twice would be registered by
// both engine generations at once, which assertNoDualServing refuses at boot; catching it here
// costs a millisecond instead of a container start and a log read.
func TestEngineSchemaNamesHasNoDuplicate(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range EngineSchemaNames() {
		if seen[name] {
			t.Errorf("schema '%s' is registered twice; a schema served by both the legacy registry "+
				"and a composable onion fails the boot in assertNoDualServing", name)
		}
		seen[name] = true
	}
}

// The count is stable across the migration: EngineSchemaNames answers every schema Sales serves,
// whichever generation serves it, so moving a resource from one to the other must not change it.
func TestEngineSchemaNamesCount(t *testing.T) {
	if got := len(EngineSchemaNames()); got != engineSchemaCount {
		t.Errorf("EngineSchemaNames() returned %d schemas, want %d; a resource was added or removed "+
			"without updating engineSchemaCount, or a cut-over added it to one list without removing "+
			"it from the other", got, engineSchemaCount)
	}
}

// The junctions are repository-only: they get a repository because a service needs to read the
// rows, but no engine, no route and no IAM resource row. transport_surface_test.go asserts the
// same thing from the other side, that a declared junction has no engine.
func TestJunctionsAreNotServedByAnEngine(t *testing.T) {
	served := map[string]bool{}
	for _, name := range EngineSchemaNames() {
		served[name] = true
	}

	for _, name := range []string{
		models.SalesChannelPaymentRelSchemaName,
		models.SalesChannelFulfillmentMethodSchemaName,
	} {
		if served[name] {
			t.Errorf("junction '%s' is served by an engine; a _rel row is configured through its "+
				"owner's capabilities, not as a CRUD resource", name)
		}
	}
}
