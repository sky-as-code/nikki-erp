package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
)

// A LangJson column keeps every translation in one jsonb document, so the text a reader sees is
// never the column itself -- it is one key out of it. These tests pin that filtering and ordering
// dive into the reader's own key, and, just as importantly, that they leave the whole-document
// behaviour exactly as it was when no locale is known.

const (
	langJsonSchemaName = "test_langjson_product"
	langJsonTagSchema  = "test_langjson_tag"
)

func langJsonSchemas(t *testing.T) (*dmodel.ModelSchema, *dmodel.SchemaRegistry) {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if registry.Get(langJsonSchemaName) != nil {
		return registry.Get(langJsonSchemaName), registry
	}

	require.NoError(t, dmodel.RegisterSchemaB(
		dmodel.DefineModel(langJsonSchemaName).
			TableName("test_langjson_products").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
			Field(dmodel.DefineField().Name("name").
				DataType(dmodel.FieldDataTypeLangJson(0, 200))).
			Field(dmodel.DefineField().Name("code").
				DataType(dmodel.FieldDataTypeString(0, 50))).
			EdgeFrom(dmodel.Edge("tags").Existing(langJsonTagSchema, "product"))))

	require.NoError(t, dmodel.RegisterSchemaB(
		dmodel.DefineModel(langJsonTagSchema).
			TableName("test_langjson_tags").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
			Field(dmodel.DefineField().Name("product_id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate()).
			Field(dmodel.DefineField().Name("label").
				DataType(dmodel.FieldDataTypeString(0, 50))).
			EdgeTo(dmodel.Edge("product").ManyToOne(
				langJsonSchemaName, dmodel.DynamicFields{"product_id": "id"}))))

	require.NoError(t, registry.FinalizeRelations())
	return registry.Get(langJsonSchemaName), registry
}

func langCode(code string) *cmodel.LanguageCode {
	c := cmodel.LanguageCode(code)
	return &c
}

// langJsonSelectSql builds the list query for one graph at one locale.
func langJsonSelectSql(
	t *testing.T, graph *dmodel.SearchGraph, language *cmodel.LanguageCode, columns []string,
) string {
	t.Helper()
	schema, registry := langJsonSchemas(t)
	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns:  ToSelectColumns(columns),
		Language: language,
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	return *sql
}

func langJsonOrderGraph(field string, dir dmodel.OrderDirection) *dmodel.SearchGraph {
	graph := dmodel.NewSearchGraph()
	graph.OrderBy(field, dir)
	return graph
}

func langJsonConditionGraph(field string, op dmodel.Operator, values ...any) *dmodel.SearchGraph {
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(field, op, values...)
	return graph
}

// --- ORDER BY ---

func TestLangJson_OrderByUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonOrderGraph("name", dmodel.Asc),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') ASC`)
}

func TestLangJson_OrderByDescUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonOrderGraph("name", dmodel.Desc),
		langCode("en-US"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'en-US') DESC`)
}

// Without a locale there is no translation to sort by, so the column is ordered as it always was.
func TestLangJson_OrderByWithoutLocaleIsUnchanged(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonOrderGraph("name", dmodel.Asc),
		nil, []string{"id", "name"})

	assert.Contains(t, sql, `ORDER BY "name" ASC`)
	assert.NotContains(t, sql, "->>")
}

// A locale must not make every column a jsonb lookup.
func TestLangJson_OrderByNonLangJsonColumnIsUnchanged(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonOrderGraph("code", dmodel.Asc),
		langCode("vi-VN"), []string{"id", "code"})

	assert.Contains(t, sql, `ORDER BY "code" ASC`)
	assert.NotContains(t, sql, "->>")
}

// Under SELECT DISTINCT every ORDER BY expression must also appear in the select list. The
// localized expression is built before the projection, so this pins that the two stay in step --
// the place this change is most likely to break.
func TestLangJson_OrderByUnderDistinctAppearsInSelectList(t *testing.T) {
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("tags.label", dmodel.Equals, "x")
	graph.OrderBy("name", dmodel.Asc)

	sql := langJsonSelectSql(t, graph, langCode("vi-VN"), []string{"id", "code"})

	assert.Contains(t, sql, "SELECT DISTINCT")
	assert.Contains(t, sql, `ASC`)
	// The localized expression is projected as well, not only sorted on.
	assert.Contains(t, sql, `->> 'vi-VN'`)
}

// --- string operators (pinning behaviour that already existed) ---

func TestLangJson_ContainsUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.Contains, "Ten"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') ILIKE E'%Ten%'`)
}

// With no locale, "contains" searches the whole document so that it still matches something.
func TestLangJson_ContainsWithoutLocaleSearchesWholeDocument(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.Contains, "Ten"),
		nil, []string{"id", "name"})

	assert.Contains(t, sql, `("name")::text ILIKE E'%Ten%'`)
}

// --- comparison operators (new behaviour) ---

func TestLangJson_EqualsUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.Equals, "TenViet"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') = E'TenViet'`)
}

func TestLangJson_NotEqualsUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.NotEquals, "Ten"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') <> E'Ten'`)
}

// "->>" yields text, so the ordering operators compare lexicographically.
func TestLangJson_GreaterThanUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.GreaterThan, "M"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') > E'M'`)
}

// The guard that keeps this change scoped: with no locale, nothing about the old path moves.
//
// That old path rejects a bare string outright -- comparing a jsonb column against one demands a
// whole document -- which is the strongest argument for localizing "=" in the first place: the
// behaviour being replaced could not express "the name reads X" at all, in any language.
func TestLangJson_EqualsWithoutLocaleStillRejectsAScalar(t *testing.T) {
	schema, registry := langJsonSchemas(t)

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(
		schema, registry, langJsonConditionGraph("name", dmodel.Equals, "Ten"),
		SqlSelectGraphOpts{Columns: ToSelectColumns([]string{"id", "name"})})

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.NotEmpty(t, *cErrs)
}

func TestLangJson_InUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.In, "Ten", "Viet"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') IN (E'Ten', E'Viet')`)
}

func TestLangJson_NotInUsesReaderLocale(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.NotIn, "Ten"),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `("name" ->> 'vi-VN') NOT IN (E'Ten')`)
}

// Null-ness is a property of the column, not of any one translation.
func TestLangJson_IsSetStaysOnTheColumn(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonConditionGraph("name", dmodel.IsNotSet),
		langCode("vi-VN"), []string{"id", "name"})

	assert.Contains(t, sql, `"name" IS NULL`)
	assert.NotContains(t, sql, "->>")
}

// A non-string operand is a client error rather than something to marshal into a document.
func TestLangJson_EqualsWithNonStringValueIsClientError(t *testing.T) {
	schema, registry := langJsonSchemas(t)

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(
		schema, registry, langJsonConditionGraph("name", dmodel.Equals, 5),
		SqlSelectGraphOpts{
			Columns:  ToSelectColumns([]string{"id", "name"}),
			Language: langCode("vi-VN"),
		})

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.NotEmpty(t, *cErrs)
}

// The locale reaches SQL as a literal, so a quote in it must not end the string. Language codes
// are validated upstream; this pins the escaping that makes that validation a second line of
// defence rather than the only one.
func TestLangJson_LocaleLiteralIsEscaped(t *testing.T) {
	sql := langJsonSelectSql(t, langJsonOrderGraph("name", dmodel.Asc),
		langCode("vi'VN"), []string{"id", "name"})

	assert.Contains(t, sql, "'vi''VN'")
}

// The count must filter exactly as the list does, or Total contradicts the rows beside it.
func TestLangJson_CountLocalizesLikeTheList(t *testing.T) {
	schema, registry := langJsonSchemas(t)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlCountGraph(
		schema, registry, langJsonConditionGraph("name", dmodel.Contains, "Ten"),
		SqlSelectGraphOpts{
			Columns:  ToSelectColumns([]string{"id", "name"}),
			Language: langCode("vi-VN"),
		})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `("name" ->> 'vi-VN') ILIKE E'%Ten%'`)
}
