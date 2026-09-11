package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// No Sales schema may declare is_archived as its own field.
//
// The column exists on every archivable resource, but it arrives through the
// core.basemodel.archivable_model mixin, and archiving is reached only through the dedicated
// actions (:id/archive, :id/unarchive, set_archived). That is what lets the module set
// RejectArchivedOnCreate for every resource at once in dynamicengines.buildOnion, the way
// Inventory does, instead of deciding it per resource.
//
// A schema that declared the field itself would make is_archived writable on create, and the
// module-wide default would start refusing a create the schema means to allow. This test is the
// guard on that assumption rather than a claim about today: it fails the moment someone adds the
// field, which is when the buildOnion decision needs revisiting.
func TestNoSchemaDeclaresIsArchivedAsItsOwnField(t *testing.T) {
	paths, err := filepath.Glob("*.json")
	if err != nil {
		t.Fatalf("listing the Sales model schemas: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no Sales model schema matched *.json; the models moved or were renamed")
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("the Sales model schema %s must be readable from the test: %v", path, err)
		}
		// The field declaration, not a mention in a label or description: only "name": "is_archived"
		// makes the field the schema's own.
		if strings.Contains(strings.ReplaceAll(string(content), " ", ""), `"name":"is_archived"`) {
			t.Errorf("%s declares is_archived as its own field; it should come from the "+
				"core.basemodel.archivable_model mixin, and declaring it makes the field writable on "+
				"create, which dynamicengines.buildOnion refuses module-wide", filepath.Base(path))
		}
	}
}
