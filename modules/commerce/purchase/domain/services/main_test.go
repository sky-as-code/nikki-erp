package services

import (
	"os"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// TestMain registers the base mixins once for the whole package: ParseModelJson panics when a named
// mixin is missing, and CoreModule.RegisterModels normally does this at start-up.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}
