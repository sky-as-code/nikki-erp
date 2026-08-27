package sales

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
)

// SALES-042's three invariants, as tests rather than as a one-time sweep.
//
// The task is "one IAM row per schema, a translation for every label, a route for every engine". Done
// once by hand that is true for an afternoon: the next schema somebody adds is unreachable over HTTP,
// or reachable and unlabelled, or routed and permission-denied — and each of those fails somewhere far
// from the omission. Checked here, the omission fails in CI next to its cause.
//
// All three parse files rather than querying a database, deliberately: they must run in CI with no
// infrastructure, and a check that needs a live system is a check that gets skipped.

// junctionSchemas are the schemas that intentionally get no engine, no route and no IAM row.
//
// A _rel row is configured through its owner's capabilities, so exposing it as a CRUD resource would
// let a client rewrite a channel's payment mapping without going through the validation that mapping
// requires. The vending_machine_new precedent is the same: 25 engines for 27 schemas.
var junctionSchemas = map[string]bool{
	"sales_channel_payment_rel": true,
}

// Every registered schema is served by an engine, except the declared junctions.
//
// A schema with no engine has no route and no IAM resource: it exists in the database and is
// unreachable, which is the failure mode hardest to notice because nothing errors — the endpoint
// simply 404s.
func TestEverySchemaHasAnEngineOrIsADeclaredJunction(t *testing.T) {
	engines := map[string]bool{}
	for _, name := range dynamicengines.EngineSchemaNames() {
		engines[name] = true
	}

	for _, schemaName := range registeredSchemaNames(t) {
		if junctionSchemas[schemaName] {
			if engines[schemaName] {
				t.Errorf("%s is declared a junction but has an engine; a _rel row is configured "+
					"through its owner's capabilities, not as a CRUD resource", schemaName)
			}
			continue
		}
		if !engines[schemaName] {
			t.Errorf("%s is registered as a schema but has no engine, so it has no route and no "+
				"IAM resource — it exists in the database and cannot be reached", schemaName)
		}
	}
}

// Every engine has an IAM resource row. Without one the engine's routes deny every request, and the
// 403 says nothing about a missing seed being the cause.
func TestEveryEngineHasAnIamResource(t *testing.T) {
	seeded := seededResourceCodes(t)

	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		if !seeded[schemaName] {
			t.Errorf("the %s engine has no iam_resources row; every route it serves will deny "+
				"every request, and the 403 will not say why", schemaName)
		}
	}
}

// Every label a schema references has a translation in every locale.
//
// A missing one surfaces as a raw key in the UI — "fields.tax_snapshot" where a column heading should
// be. Harmless in isolation, and the reason nobody notices until a customer does.
func TestEveryLabelIsTranslated(t *testing.T) {
	labels := referencedLabels(t)
	if len(labels) == 0 {
		t.Fatal("no labels parsed out of the schemas; the parser or the files changed")
	}

	for _, locale := range []string{"en-US", "vi-VN"} {
		translations := translationsFor(t, locale)

		missing := make([]string, 0)
		for label := range labels {
			if _, ok := translations[label]; !ok {
				missing = append(missing, label)
			}
		}
		sort.Strings(missing)

		if len(missing) > 0 {
			t.Errorf("%s is missing %d of %d labels: %v",
				locale, len(missing), len(labels), missing)
		}
	}
}

// Labels are snake_case, matching the field names they describe.
//
// A camelCase key is not wrong so much as unfindable: somebody adding a translation looks for the
// field's own name, does not find it, and adds a second key beside the first. One such typo
// (fields.taxSnapshot) is exactly how this test came to exist.
func TestLabelKeysAreSnakeCase(t *testing.T) {
	camel := regexp.MustCompile(`[a-z][A-Z]`)

	for label := range referencedLabels(t) {
		if camel.MatchString(label) {
			t.Errorf("the label %q is camelCase; keys are snake_case so that somebody looking for "+
				"a field's translation finds it under the field's own name", label)
		}
	}
}

// registeredSchemaNames reads the schema names out of the module's RegisterModels list.
//
// Parsed from the source rather than by calling RegisterModels, which would need the base schemas
// registered and a container built — infrastructure this test deliberately does without.
func registeredSchemaNames(t *testing.T) []string {
	t.Helper()

	models := readAllGoFiles(t, filepath.Join("domain", "models"))
	byBuilder := map[string]string{}
	for _, match := range regexp.MustCompile(
		`(Sales\w+?)SchemaName\s*=\s*"([a-z_]+)"`).FindAllStringSubmatch(models, -1) {
		byBuilder[match[1]] = match[2]
	}

	indexBytes, err := os.ReadFile("index.go")
	if err != nil {
		t.Fatalf("reading index.go: %v", err)
	}
	index := string(indexBytes)
	names := make([]string, 0, len(byBuilder))
	for _, match := range regexp.MustCompile(
		`models\.(Sales\w+?)SchemaBuilder`).FindAllStringSubmatch(index, -1) {
		if schemaName, ok := byBuilder[match[1]]; ok {
			names = append(names, schemaName)
		} else {
			t.Errorf("%sSchemaBuilder is registered but declares no SchemaName constant", match[1])
		}
	}
	if len(names) == 0 {
		t.Fatal("no schemas parsed out of index.go; the parser or the file changed")
	}
	return names
}

// seededResourceCodes reads the iam_resources codes out of every Sales IAM migration.
func seededResourceCodes(t *testing.T) map[string]bool {
	t.Helper()

	dir := filepath.Join("..", "..", "scripts", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the migrations directory: %v", err)
	}

	codes := map[string]bool{}
	pattern := regexp.MustCompile(`'(sales_[a-z_]+)'`)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.Contains(name, "sales") || !strings.Contains(name, "iam") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		// The resource code is the third quoted field of an iam_resources row. Matching every
		// sales_* literal over-collects — action descriptions mention them too — which is the safe
		// direction: a false positive here weakens the test, a false negative would fail the build
		// over a seed that exists.
		for _, match := range pattern.FindAllStringSubmatch(string(content), -1) {
			codes[match[1]] = true
		}
	}
	if len(codes) == 0 {
		t.Fatal("no resource codes parsed out of the IAM migrations")
	}
	return codes
}

// referencedLabels collects every label key the schemas point at.
func referencedLabels(t *testing.T) map[string]bool {
	t.Helper()

	dir := filepath.Join("domain", "models")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the models directory: %v", err)
	}

	labels := map[string]bool{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}

		var schema struct {
			Label  string `json:"label"`
			Fields []struct {
				Label string `json:"label"`
			} `json:"fields"`
		}
		if err := json.Unmarshal(content, &schema); err != nil {
			t.Fatalf("parsing %s: %v", entry.Name(), err)
		}

		if schema.Label != "" {
			labels[schema.Label] = true
		}
		for _, field := range schema.Fields {
			if field.Label != "" {
				labels[field.Label] = true
			}
		}
	}
	return labels
}

// translationsFor reads one locale's Sales translation pack.
func translationsFor(t *testing.T, locale string) map[string]any {
	t.Helper()

	path := filepath.Join("..", "essential", "infra", "langJson", locale, "sales.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the %s translations: %v", locale, err)
	}

	var translations map[string]any
	if err := json.Unmarshal(content, &translations); err != nil {
		t.Fatalf("parsing the %s translations: %v", locale, err)
	}
	return translations
}

func readAllGoFiles(t *testing.T, dir string) string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}

	var builder strings.Builder
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		builder.Write(content)
	}
	return builder.String()
}
