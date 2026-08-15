package computed_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// The DSL's injection surface, attacked seam by seam. The framework's core defense is that every
// identifier in a definition must resolve through the schema registry at finalize time, and
// every literal is evaluated in Go, never rendered into SQL. These tests prove a hostile
// definition dies at validation, and hostile VALUES that do reach SQL (the batched IN read) go
// through the query builder's escaping.
//
// Deferred to the SQL-based phase (aggregate/exists/lookup, stored computed fields): fan-out
// prevention, computed-field filter/sort/pagination, and query-context tests — those features do
// not exist in the Go-executed phase, so there is nothing to attack yet.

// hostileNames are identifier payloads an attacker might plant in a definition.
func hostileNames() []string {
	return []string{
		`name";DROP TABLE users;--`,
		`name'; DELETE FROM x; --`,
		`name" OR "1"="1`,
		"name\x00hidden",
		"name​zero_width",
		`../secret`,
		`na"me`,
		`na'me`,
	}
}

func TestInjection_HostileFieldNamesFailResolution(t *testing.T) {
	for i, hostile := range hostileNames() {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			schema := dmodel.DefineModel(fmt.Sprintf("cf_inj_field_%d", i)).
				Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
				Field(dmodel.DefineField().Name("evil").DataType(dmodel.FieldDataTypeString(0, 100)).
					Computed(false, computed.Fn("lower", computed.F(hostile)))).
				Build()
			reg := dmodel.NewSchemaRegistry()
			require.NoError(t, reg.Register(schema))

			err := reg.FinalizeRelations()
			require.Error(t, err, "hostile operand %q must fail registry resolution", hostile)
		})
	}
}

func TestInjection_HostileRelatedPathsFailResolution(t *testing.T) {
	paths := append(hostileNames(),
		`template."name`,
		`template.name;DROP TABLE x`,
		`template..name`,
		`.name`,
		`template.`,
		`template.name.`,
	)
	for i, hostile := range paths {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			reg := newRegistryWith(t, templateLikeSchema(), variantLikeSchema(
				dmodel.DefineField().Name("evil").DataType(dmodel.FieldDataTypeString(0, 200)).
					Computed(false, computed.Related(hostile)),
			))

			err := reg.FinalizeRelations()
			require.Error(t, err, "hostile related path %q must fail resolution", hostile)
		})
	}
}

func TestInjection_HostileJsonDefinitionsRejected(t *testing.T) {
	template := `{
		"name": "cf_inj_json",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true},
			{"name": "evil", "data_type": {"type": "string", "min": 0, "max": 100}, "computed": %s}
		]
	}`
	blocks := map[string]string{
		"raw sql property": `{"kind": "expression", "is_stored": false,
			"sql": "SELECT pg_sleep(10)",
			"expression": {"field": "id"}}`,
		"table name smuggled as function": `{"kind": "expression", "is_stored": false,
			"expression": {"function": "pg_read_file", "args": [{"value": "/etc/passwd"}]}}`,
		"subquery in value is inert but object-typed": `{"kind": "expression", "is_stored": false,
			"expression": {"value": {"$subquery": "SELECT 1"}}}`,
	}
	for name, block := range blocks {
		t.Run(name, func(t *testing.T) {
			modelJson := fmt.Sprintf(template, block)
			_, errs := dmodel.ParseModelJsonSafe(modelJson)
			if errs.Count() > 0 {
				return // rejected by the JSON Schema — first line of defense
			}
			// Parsed structurally (e.g. an unknown function name is schema-legal); it must then
			// die at finalize, never reaching evaluation or SQL.
			schema := dmodel.ParseModelJson(modelJson).Build()
			reg := dmodel.NewSchemaRegistry()
			require.NoError(t, reg.Register(schema))
			require.Error(t, reg.FinalizeRelations())
		})
	}
}

// The projection augmentation seam: whatever a definition contains, the only names the eval plan
// may append to a request's field list are operands that resolved through the schema registry.
func TestInjection_ExtraFieldsOnlyContainSchemaFields(t *testing.T) {
	schema := finalizeChainFixture(t)

	plan, errs := computed.BuildEvalPlan("cf_eval_chain", []string{"total"})
	require.Equal(t, 0, errs.Count())
	require.NotNil(t, plan)

	for _, name := range plan.ExtraFields {
		_, ok := schema.Field(name)
		assert.True(t, ok, "appended projection name %q must be a registered schema field", name)
	}
}

// The one seam where attacker-controlled VALUES genuinely reach SQL: the batched IN(...) source
// read keys come from row data. They must go through the query builder's escaping — this asserts
// hostile key values end up quoted, with their quotes doubled, never as executable SQL.
func TestInjection_BatchedReadKeysAreEscaped(t *testing.T) {
	schema := dmodel.DefineModel("cf_inj_esc").
		ShouldBuildDb().
		TableName("cf_inj_esc").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(schema))
	require.NoError(t, reg.FinalizeRelations())

	hostileKey := `x'; DROP TABLE cf_inj_esc; --`
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("id", dmodel.In, hostileKey, "t2")

	sql, clientErrs, err := orm.NewPgQueryBuilder().SqlSelectGraph(
		schema, reg, graph, orm.SqlSelectGraphOpts{Columns: orm.ToSelectColumns([]string{"id", "name"})})
	require.NoError(t, err)
	require.Nil(t, clientErrs)
	require.NotNil(t, sql)

	// The builder renders values as E'...' literals with backslash-escaped quotes, so the
	// payload's terminating quote must appear only in escaped form: never `'x'; DROP ...`.
	assert.Contains(t, *sql, `E'x\'; DROP TABLE cf_inj_esc; --'`,
		"the payload must survive only as an escaped string literal")
	assert.NotContains(t, *sql, `'x'; DROP`,
		"an unescaped quote before the payload would terminate the literal and execute the rest")
	assert.Equal(t, 1, strings.Count(*sql, "DROP TABLE"), "the payload appears once, inside its literal")
}

// Expression literals never touch SQL at all: the whole expression evaluates in Go over the row.
func TestInjection_ExpressionLiteralsStayInGo(t *testing.T) {
	payload := `'; DROP TABLE x; --`
	expr := computed.Fn("concat", computed.Lit("a"), computed.Lit(payload))

	got, err := computed.Eval(expr, dmodel.DynamicFields{})
	require.NoError(t, err)
	assert.Equal(t, "a"+payload, got, "the literal is just a Go string value, produced verbatim")
}
