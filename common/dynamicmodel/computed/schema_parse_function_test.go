package computed

import (
	"encoding/json"
	"strings"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func TestFunctionKindJsonRoundTrip(t *testing.T) {
	raw := []byte(`{"kind":"function","is_stored":false,"function":"inventory.effective_sales_tax_ids","depends_on":"sales_tax_mode"}`)
	expr, isStored, err := ParseDefinitionJson(raw, "effective_sales_tax_ids")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if isStored {
		t.Fatal("is_stored must be false")
	}
	def, err := NewDefinition(isStored, expr)
	if err != nil {
		t.Fatalf("definition: %v", err)
	}
	if def.Kind != ComputeFunction {
		t.Fatalf("kind = %q, want function", def.Kind)
	}
	if def.Function.Name != "inventory.effective_sales_tax_ids" || def.Function.DependsOn != "sales_tax_mode" {
		t.Fatalf("bad node: %+v", def.Function)
	}
}

func TestFunctionKindRejectsEmptyName(t *testing.T) {
	_, _, err := ParseDefinitionJson([]byte(`{"kind":"function","is_stored":false}`), "x")
	if err == nil {
		t.Fatal("expected an error for a missing function name")
	}
}

func TestFunctionKindIsRootOnly(t *testing.T) {
	nested := Add(GoFunction("f").Build(), Lit(int64(1)))
	if _, err := NewDefinition(false, nested); err == nil {
		t.Fatal("expected a root-only rejection")
	}
}

func TestChainedBuilderMatchesJson(t *testing.T) {
	def, err := NewDefinition(false, GoFunction("f").DependsOn("mode").Build())
	if err != nil {
		t.Fatalf("chained: %v", err)
	}
	if def.Kind != ComputeFunction || def.Function.DependsOn != "mode" {
		t.Fatalf("bad def: %+v", def)
	}
}

// The frontend cannot know a field is function-computed, or what to watch for a recompute, unless
// meta/schema says so.
func TestToSimplizedCarriesComputedDescriptor(t *testing.T) {
	schema := dmodel.DefineModel("cf_desc").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("mode").DataType(dmodel.FieldDataTypeString(0, 20))).
		Field(dmodel.DefineField().Name("tag_ids").DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Computed(false, GoFunction("tags").DependsOn("mode").Build())).
		Build()

	encoded, err := json.Marshal(schema.ToSimplized())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `"kind":"function"`) {
		t.Fatalf("descriptor kind missing from: %s", text)
	}
	if !strings.Contains(text, `"depends_on":"mode"`) {
		t.Fatalf("depends_on missing from: %s", text)
	}
}

// A non-function computed field reports its kind but has nothing to watch.
func TestToSimplizedOmitsDependsOnForOtherKinds(t *testing.T) {
	schema := dmodel.DefineModel("cf_desc_expr").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt64(0, 1000))).
		Field(dmodel.DefineField().Name("doubled").DataType(dmodel.FieldDataTypeInt64(0, 2000)).
			Computed(false, Mul(F("qty"), Lit(int64(2))))).
		Build()

	encoded, _ := json.Marshal(schema.ToSimplized())
	text := string(encoded)
	if !strings.Contains(text, `"kind":"expression"`) {
		t.Fatalf("expression kind missing from: %s", text)
	}
	if strings.Contains(text, "depends_on") {
		t.Fatalf("depends_on must be omitted for non-function kinds: %s", text)
	}
}
