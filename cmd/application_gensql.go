package cmd

import (
	"fmt"
	"os"
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules"
)

// GenSql renders the module's CREATE TABLE statements. withDeps additionally renders the
// tables of every module in the dependency closure, which a cross-module foreign key needs:
// the constraint names a table this module does not own, and Atlas cannot resolve a reference
// to a table absent from the schema it is given. ForEachOrder is dependency-ordered, so the
// dependency tables come first.
func (this *Application) GenSql(moduleName string, dialect string, withDeps bool) string {
	registry, prefixes, registeredCount := this.prepareSqlGen(moduleName)
	genPrefixes := prefixes
	if withDeps {
		// No prefix filter: take everything registered, which registerModelsUpTo limited to
		// this module plus its transitive dependencies.
		genPrefixes = nil
	}
	queries := mustGenCreateSql(registry, dialect, genPrefixes)

	// Schemas registered but none matched the prefixes: ModelPrefixes() is wrong for this module.
	if len(queries) == 0 && registeredCount > 0 {
		fmt.Fprintf(os.Stderr,
			"module '%s' registered %d schema(s) but none match '%s'; "+
				"fix ModelPrefixes() in the module's index.go\n",
			moduleName, registeredCount, strings.Join(prefixes, "', '"))
		os.Exit(1)
	}
	return strings.Join(queries, ";\n")
}

// GenDepsSql renders only the dependency closure's statements: GenSql(withDeps=true) minus
// the module's own. The migration script writes them as a baseline migration, so Atlas sees the
// dependency tables as already existing and diffs this module alone.
//
// A baseline rather than Atlas's `exclude`: excluding a table also drops every foreign key
// that references it, which is every cross-module foreign key.
func (this *Application) GenDepsSql(moduleName string, dialect string) string {
	registry, prefixes, _ := this.prepareSqlGen(moduleName)
	own := make(map[string]struct{})
	for _, query := range mustGenCreateSql(registry, dialect, prefixes) {
		own[query] = struct{}{}
	}

	depQueries := make([]string, 0)
	for _, query := range mustGenCreateSql(registry, dialect, nil) {
		if _, isOwn := own[query]; !isOwn {
			depQueries = append(depQueries, query)
		}
	}
	if len(depQueries) == 0 {
		return ""
	}
	return strings.Join(depQueries, ";\n") + ";\n"
}

// prepareSqlGen registers the module's dependency closure and finalizes it, returning the
// registry, the module's schema prefixes (each with its "_" separator) and how many schemas the module itself registered.
func (this *Application) prepareSqlGen(moduleName string) (*dmodel.SchemaRegistry, []string, int) {
	registeredCount, err := this.registerModelsUpTo(moduleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to register models: %v\n", err)
		os.Exit(1)
	}

	registry := dmodel.GetSchemaRegistry()
	if err := registry.FinalizeRelations(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to finalize schema relations: %v\n", err)
		os.Exit(1)
	}
	if err := orm.ValidateRelations(registry); err != nil {
		fmt.Fprintf(os.Stderr, "failed to validate relations: %v\n", err)
		os.Exit(1)
	}

	modelPrefixes := this.buildModuleMap()[moduleName].(modules.DynamicModule).ModelPrefixes()
	prefixes := make([]string, 0, len(modelPrefixes))
	for _, prefix := range modelPrefixes {
		prefixes = append(prefixes, prefix+"_")
	}
	return registry, prefixes, registeredCount
}

func mustGenCreateSql(registry *dmodel.SchemaRegistry, dialect string, schemaPrefixes []string) []string {
	queries, err := orm.GenCreateSql(registry, dialect, schemaPrefixes...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate create SQL: %v\n", err)
		os.Exit(1)
	}
	return queries
}
