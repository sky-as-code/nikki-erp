package services

import (
	"os"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// TestMain registers the base mixins once for the whole package.
//
// Building any schema here parses its JSON, and ParseModelJson panics when a named mixin is not in
// the registry — normally CoreModule.RegisterModels does this during start-up.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}
