package usagecheck

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// ReferencingTable names one column that may hold a reference to the checked resource.
type ReferencingTable struct {
	SchemaName string
	Field      string
}

// RepositoryResolver answers which repository serves a schema. Every module has one - a resource
// hub, or the engine registry - and it is resolved at check time rather than at registration,
// because checkers register during Init while the onions are built later.
type RepositoryResolver func(schemaName string) (RowSearcher, error)

// RowSearcher is the single method a usage check needs. Narrowed to one method so a checker can
// be unit-tested against a stub instead of a built engine.
type RowSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// NewTableChecker answers "is this resource referenced" by looking for a single row in each of
// the given tables. It is the common shape of every dependant's checker; a module whose answer
// depends on business state beyond the reference itself writes its own instead.
//
// Archived rows count. An archived order's lines are history, and history is exactly what must
// not be reinterpreted by deleting the unit or product it names.
//
// A lookup that fails returns an error rather than false: the caller blocks a delete on an error
// and permits it on a false, so the two must never be confused.
func NewTableChecker(resolve RepositoryResolver, tables ...ReferencingTable) ResourceUsageChecker {
	return CheckerFunc(func(ctx corectx.Context, ref ResourceRef) (bool, string, error) {
		id := ref.Identifier[IdentifierKeyId]
		if id == "" {
			return false, "", nil
		}

		for _, table := range tables {
			found, err := anyRowReferencing(ctx, resolve, table, id, ref.Identifier[IdentifierKeyOrgId])
			if err != nil {
				return false, "", err
			}
			if found {
				return true, table.SchemaName, nil
			}
		}
		return false, "", nil
	})
}

// anyRowReferencing reports whether one table holds at least one row pointing at the resource.
func anyRowReferencing(
	ctx corectx.Context, resolve RepositoryResolver, table ReferencingTable, id string, orgId string,
) (bool, error) {
	repo, err := resolve(table.SchemaName)
	if err != nil {
		return false, errors.Wrapf(err, "resolving %s to check references to %s", table.SchemaName, id)
	}

	// Every condition goes into ONE And call: SearchGraph.And REPLACES its node list rather than
	// appending, so a second call would silently drop the first condition — leaving a search that
	// matches every row in the table and reports any resource as used.
	conditions := []dmodel.SearchNode{
		*dmodel.NewSearchNode().NewCondition(table.Field, dmodel.Equals, id),
	}
	// Organization scope, when the checked resource has one and this table carries it: a
	// reference held by another organization is that organization's business, and treating it
	// as usage here would both block the wrong delete and disclose that the row exists.
	if orgId != "" && schemaHasField(repo, IdentifierKeyOrgId) {
		conditions = append(conditions,
			*dmodel.NewSearchNode().NewCondition(IdentifierKeyOrgId, dmodel.Equals, orgId))
	}

	graph := &dmodel.SearchGraph{}
	graph.And(conditions...)

	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		// The question is "any", so a second row costs a page for an answer already settled.
		Size: 1,
		// nil rather than false: archived rows are still references.
		IncludeArchived: nil,
	})
	if err != nil {
		return false, errors.Wrapf(err, "checking whether %s.%s references %s",
			table.SchemaName, table.Field, id)
	}
	return found != nil && found.HasData && len(found.Data.Items) > 0, nil
}

// schemaHasField reports whether the searched resource carries the column, so that a table
// without org scoping is not filtered by a column it does not have.
func schemaHasField(repo RowSearcher, field string) bool {
	schemaHolder, ok := repo.(interface{ Schema() *dmodel.ModelSchema })
	if !ok {
		return false
	}
	schema := schemaHolder.Schema()
	if schema == nil {
		return false
	}
	_, found := schema.Field(field)
	return found
}
