package sales

import (
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// TestRegisterModelsParsesEverySchema boots the module's real registration path, because `go build`
// cannot see inside an embedded JSON file and a malformed schema would otherwise panic during
// application boot. It drives RegisterModels rather than a list of builders so a newly added schema
// cannot be omitted, and so that a schema declared before the one its edge points at — which parses
// fine in isolation but fails to register — is also caught.
func TestRegisterModelsParsesEverySchema(t *testing.T) {
	// Sales schemas extend the core.basemodel.* schemas; without these, ParseModelJson panics on
	// the first `extend_before` it cannot resolve.
	_ = basemodel.RegisterJsonBaseSchemas()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterModels panicked, so a schema is malformed or declared before the "+
				"schema its edge points at: %v", r)
		}
	}()

	if err := ModuleSingleton.RegisterModels(); err != nil {
		t.Fatalf("RegisterModels: %v", err)
	}
}
