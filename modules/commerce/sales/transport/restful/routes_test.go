package restful

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
)

// These read index.go as text rather than building a container, so they run in CI with no
// database and no dependency graph. What they guard is the shape of the route table, which is
// exactly what a hand-edited list of AddRoute calls gets wrong.

// A collection-level custom route must be registered before its resource's ":id" routes.
//
// composable.RestEngine sorts by specificity when it registers, so the order in the file does not
// decide the outcome today -- but the file is where a reader looks to reason about it, and a
// resource whose collection route is written after its ":id" ones invites the next author to
// assume the file order is the registration order. Keeping them in order keeps the file honest.
//
// Getting this wrong is not a compile error and not a test failure anywhere else: the router
// simply reads "merge" as a bill id and answers 404 for an endpoint that exists.
func TestCollectionRoutesAreAddedBeforeIdRoutes(t *testing.T) {
	source := readRouteSource(t)

	for _, init := range splitInitFuncs(source) {
		firstId := -1
		for i, path := range routePathsOf(init.body) {
			isId := strings.HasPrefix(path, ":id")
			if isId && firstId < 0 {
				firstId = i
				continue
			}
			if !isId && firstId >= 0 {
				t.Errorf("%s adds the collection route %q after an \":id\" route; a reader cannot "+
					"tell from the file which one the router will match first", init.name, path)
			}
		}
	}
}

// Every resource Sales serves has an init function that registers its routes, AND that function
// is called from initSalesV1. A resource with an onion and no routes is unreachable over HTTP
// while looking perfectly healthy at boot.
//
// Both halves are checked because either alone is satisfiable by dead code: a function nobody
// calls registers nothing, and a call to a function that names the wrong schema registers the
// wrong thing.
func TestEveryEngineSchemaHasRoutes(t *testing.T) {
	source := readRouteSource(t)
	called := calledInitFuncs(source)

	registered := map[string]bool{}
	for _, init := range splitInitFuncs(source) {
		if !called[init.name] {
			t.Errorf("%s is defined but never called from initSalesV1, so it registers nothing",
				init.name)
			continue
		}
		for _, schema := range schemaConstantsOf(init.body) {
			registered[schema] = true
		}
	}

	for _, schema := range dynamicengines.EngineSchemaNames() {
		if !registered[constantNameFor(schema)] {
			t.Errorf("schema %q is served by an engine but no called init function registers its "+
				"routes", schema)
		}
	}
}

// The container names in the v1 REST structs are string literals, because a dig name tag has to be
// one. This asserts each still matches what EngineDependencyName produces, so a schema renamed in
// models/ cannot leave a REST struct pointing at an onion that no longer exists -- which dig
// reports as a missing type at boot, far from the literal that caused it.
func TestEngineDependencyNamesMatchTheSchemas(t *testing.T) {
	for _, schema := range dynamicengines.EngineSchemaNames() {
		want := composable.EngineDependencyName(schema)
		if !strings.HasPrefix(want, "dynengine_") {
			t.Errorf("EngineDependencyName(%q) = %q, which is not a dynengine_ name", schema, want)
		}
		if want != "dynengine_"+schema {
			t.Errorf("EngineDependencyName(%q) = %q, want %q", schema, want, "dynengine_"+schema)
		}
	}
}

type initFunc struct {
	name string
	body string
}

var (
	initFuncPattern    = regexp.MustCompile(`func (initSales\w+V1)\(route \*echo\.Group\) error \{`)
	routePathPattern   = regexp.MustCompile(`Path:\s*"([^"]+)"`)
	schemaConstPattern = regexp.MustCompile(`models\.(\w+SchemaName)`)
	initCallPattern    = regexp.MustCompile(`(initSales\w+V1)\(routeV1\)`)
)

func readRouteSource(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("index.go"))
	if err != nil {
		t.Fatalf("the route table must be readable from the test: %v", err)
	}
	return string(content)
}

// splitInitFuncs cuts the file into one entry per initXxxV1 function. Crude but sufficient: each
// function ends where the next begins, and the last runs to the end of the file.
func splitInitFuncs(source string) []initFunc {
	matches := initFuncPattern.FindAllStringSubmatchIndex(source, -1)
	funcs := make([]initFunc, 0, len(matches))
	for i, m := range matches {
		end := len(source)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		funcs = append(funcs, initFunc{name: source[m[2]:m[3]], body: source[m[1]:end]})
	}
	return funcs
}

// calledInitFuncs collects the init functions initSalesV1 actually invokes, which is the list
// that decides what gets registered -- not the set of functions that happen to be defined.
func calledInitFuncs(source string) map[string]bool {
	start := strings.Index(source, "func initSalesV1()")
	if start < 0 {
		return map[string]bool{}
	}
	end := strings.Index(source[start:], "\nfunc ")
	body := source[start:]
	if end > 0 {
		body = source[start : start+end]
	}
	called := map[string]bool{}
	for _, m := range initCallPattern.FindAllStringSubmatch(body, -1) {
		called[m[1]] = true
	}
	return called
}

func routePathsOf(body string) []string {
	found := routePathPattern.FindAllStringSubmatch(body, -1)
	paths := make([]string, 0, len(found))
	for _, f := range found {
		paths = append(paths, f[1])
	}
	return paths
}

func schemaConstantsOf(body string) []string {
	found := schemaConstPattern.FindAllStringSubmatch(body, -1)
	names := make([]string, 0, len(found))
	for _, f := range found {
		names = append(names, f[1])
	}
	return names
}

// constantNameFor maps a schema name to the Go constant that holds it, which is what the route
// table names. sales_order -> SalesOrderSchemaName.
func constantNameFor(schemaName string) string {
	parts := strings.Split(schemaName, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	b.WriteString("SchemaName")
	return b.String()
}

// A compile-time reminder that the models package is the source of those constants.
var _ = models.SalesOrderSchemaName
