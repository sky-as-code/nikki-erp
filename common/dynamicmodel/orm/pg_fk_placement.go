package orm

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// fkPlacement moves a foreign key out of its CREATE TABLE when the referenced table is created
// later in the same script. Tables are created in FK-dependency order, but a cycle (two tables
// whose edges reference each other) has no such order: whichever table comes first would
// reference one that is not created yet. Those constraints become
// `ALTER TABLE ... ADD CONSTRAINT` statements emitted after every CREATE TABLE.
//
// A nil placement inlines every constraint, which is what a single-table SqlCreateTable wants.
type fkPlacement struct {
	created  map[string]bool
	deferred []string
}

func newFkPlacement() *fkPlacement {
	return &fkPlacement{created: make(map[string]bool)}
}

func (this *fkPlacement) markCreated(schemaName string) {
	this.created[schemaName] = true
}

func (this *fkPlacement) mustDefer(owner, ref *dmodel.ModelSchema) bool {
	if this == nil || owner.Name() == ref.Name() {
		return false
	}
	return !this.created[ref.Name()]
}

func (this *fkPlacement) deferConstraint(owner *dmodel.ModelSchema, quotedName, body string) {
	this.deferred = append(this.deferred, fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY %s", pgQuote(owner.TableName()), quotedName, body))
}
