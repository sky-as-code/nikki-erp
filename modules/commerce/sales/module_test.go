package sales

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOwnAppServicesAreNotBoundAsExternalPorts guards an init-ordering cycle that only shows up at
// boot: infra/external binds ports eagerly and runs first in Init, while Sales' own application
// services register several steps later, so resolving one of them in InitExternal fails the boot
// while every unit test still passes.
func TestOwnAppServicesAreNotBoundAsExternalPorts(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "infra", "sales", "external", "index.go"))
	if err != nil {
		t.Fatalf("commerce/infra/sales/external/index.go must be readable: %v", err)
	}

	// Sales' own interfaces packages. A port bound in infra/external must belong to ANOTHER module.
	for _, own := range []string{
		"modules/commerce/sales/interfaces/channel",
	} {
		if strings.Contains(string(source), own) {
			t.Errorf("infra/external binds %q, which is one of Sales' own application services; "+
				"it is registered after InitExternal runs, so the container cannot resolve it "+
				"there and Init will fail at boot", own)
		}
	}
}
