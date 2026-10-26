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

// Every schema Sales serves appears exactly once. A schema listed twice would register its onion
// and its routes twice; catching it here costs a millisecond instead of a container start.
func TestEngineSchemaNamesHasNoDuplicate(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range EngineSchemaNames() {
		if seen[name] {
			t.Errorf("schema '%s' is registered twice", name)
		}
		seen[name] = true
	}
}

// EngineSchemaNames answers every schema Sales serves through an onion.
func TestEngineSchemaNamesCount(t *testing.T) {
	if got := len(EngineSchemaNames()); got != engineSchemaCount {
		t.Errorf("EngineSchemaNames() returned %d schemas, want %d; a resource was added or removed "+
			"without updating engineSchemaCount", got, engineSchemaCount)
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
