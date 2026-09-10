package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// The decorator generalizes what modules used to hand-roll for template_*-style fields, so these
// tests pin the behaviours those hand-rolled overrides guaranteed: one batched source query per
// page, the FK forced into an explicit projection, distinct keys fetched once, a missing source
// row leaving fields absent, and untouched requests passing straight through.

// computedBaseService is the wrapped domain service: it records the params it receives and
// returns a canned page.
type computedBaseService struct {
	CrudDomainService
	schema     *dmodel.ModelSchema
	gotParams  dmodel.DynamicFields
	searchPage []dmodel.DynamicFields
	single     dmodel.DynamicFields
	calls      int
}

func (this *computedBaseService) Schema() *dmodel.ModelSchema { return this.schema }

func (this *computedBaseService) Search(
	_ corectx.Context, params SearchQuery, _ ...corecrud.ServiceSearchOptions,
) (*SearchResult, error) {
	this.calls++
	this.gotParams = params
	return &SearchResult{Data: dyn.PagedResultData[dmodel.DynamicFields]{Items: this.searchPage}, HasData: true}, nil
}

func (this *computedBaseService) GetById(_ corectx.Context, params GetByIdQuery) (*GetOneResult, error) {
	this.calls++
	this.gotParams = params
	return &GetOneResult{Data: dyn.SingleResultData[dmodel.DynamicFields]{Item: this.single}, HasData: true}, nil
}

func (this *computedBaseService) Create(_ corectx.Context, params CreateCommand, _ ...CreateOptions) (*CreateResult, error) {
	this.calls++
	return &CreateResult{Data: params, HasData: true}, nil
}

func (this *computedBaseService) Update(_ corectx.Context, _ UpdateCommand, _ ...UpdateOptions) (*MutateResult, error) {
	this.calls++
	return &MutateResult{HasData: true}, nil
}

// sourceSearchRecorder is the batched-read seam, returning canned source rows.
type sourceSearchRecorder struct {
	calls      int
	gotSchema  string
	gotKeys    []any
	gotFields  []string
	sourceRows []dmodel.DynamicFields
}

func (this *sourceSearchRecorder) fn() SourceSearchFn {
	return func(
		_ corectx.Context, schemaName string, _ string, keys []any, fields []string,
	) ([]dmodel.DynamicFields, error) {
		this.calls++
		this.gotSchema, this.gotKeys, this.gotFields = schemaName, keys, fields
		return this.sourceRows, nil
	}
}

func buildDecoratorFixture(t *testing.T) (*computedBaseService, *sourceSearchRecorder, CrudDomainService) {
	return buildDecoratorFixtureWithDefaults(t, nil)
}

func buildDecoratorFixtureWithDefaults(
	t *testing.T, defaultSearchFields []string,
) (*computedBaseService, *sourceSearchRecorder, CrudDomainService) {
	t.Helper()
	source := dmodel.DefineModel("cmp_cf_template").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
	owner := dmodel.DefineModel("cmp_cf_variant").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name"))).
		EdgeTo(dmodel.Edge("template").ManyToOne("cmp_cf_template", dmodel.DynamicFields{"template_id": "id"})).
		Build()

	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(source))
	require.NoError(t, reg.Register(owner))
	require.NoError(t, reg.FinalizeRelations())

	base := &computedBaseService{schema: owner}
	recorder := &sourceSearchRecorder{}
	// nil invoker: this fixture declares no function-kind fields, so one is never called.
	return base, recorder, WithComputedFields(base, recorder.fn(), nil, defaultSearchFields)
}

func TestComputedServiceSearchBatchesOneSourceQuery(t *testing.T) {
	base, recorder, service := buildDecoratorFixture(t)
	base.searchPage = []dmodel.DynamicFields{
		{"id": "v1", "template_id": "t1"},
		{"id": "v2", "template_id": "t2"},
		{"id": "v3", "template_id": "t1"}, // duplicate key fetched once
		{"id": "v4", "template_id": nil},
		{"id": "v5", "template_id": "t9"}, // dead reference
	}
	recorder.sourceRows = []dmodel.DynamicFields{
		{"id": "t1", "name": "Widget"},
		{"id": "t2", "name": "Gadget"},
	}

	result, err := service.Search(nil, dmodel.DynamicFields{"fields": []string{"id", "template_name"}})
	require.NoError(t, err)

	assert.Equal(t, 1, base.calls, "the page itself costs one query")
	assert.Equal(t, 1, recorder.calls, "the source read costs exactly one more, however large the page")
	assert.Equal(t, "cmp_cf_template", recorder.gotSchema)
	assert.ElementsMatch(t, []any{"t1", "t2", "t9"}, recorder.gotKeys)
	assert.ElementsMatch(t, []string{"id", "name"}, recorder.gotFields)

	rows := result.Data.Items
	assert.Equal(t, "Widget", rows[0]["template_name"])
	assert.Equal(t, "Gadget", rows[1]["template_name"])
	assert.Equal(t, "Widget", rows[2]["template_name"])
	_, present := rows[3]["template_name"]
	assert.False(t, present, "no key means no value, not an empty one")
	_, present = rows[4]["template_name"]
	assert.False(t, present, "a dead reference reads as unknown")
}

func TestComputedServiceForcesFkIntoExplicitProjection(t *testing.T) {
	base, _, service := buildDecoratorFixture(t)
	base.searchPage = []dmodel.DynamicFields{}

	_, err := service.Search(nil, dmodel.DynamicFields{"fields": []string{"id", "template_name"}})
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"id", "template_name", "template_id"}, base.gotParams["fields"],
		"the FK join key must ride along, or the rows cannot be matched back to their sources")
}

func TestComputedServiceUntouchedWhenNothingComputedRequested(t *testing.T) {
	base, recorder, service := buildDecoratorFixture(t)
	base.searchPage = []dmodel.DynamicFields{{"id": "v1", "template_id": "t1"}}

	_, err := service.Search(nil, dmodel.DynamicFields{"fields": []string{"id", "template_id"}})
	require.NoError(t, err)

	assert.Equal(t, 0, recorder.calls, "no computed field requested, no source read")
	assert.Equal(t, []string{"id", "template_id"}, base.gotParams["fields"])
}

func TestComputedServiceGetByIdFillsSingleRecord(t *testing.T) {
	base, recorder, service := buildDecoratorFixture(t)
	base.single = dmodel.DynamicFields{"id": "v1", "template_id": "t1"}
	recorder.sourceRows = []dmodel.DynamicFields{{"id": "t1", "name": "Widget"}}

	result, err := service.GetById(nil, dmodel.DynamicFields{"fields": []string{"id", "template_name"}})
	require.NoError(t, err)
	assert.Equal(t, "Widget", result.Data.Item["template_name"])
}

func TestComputedServiceWritesToComputedFieldRejected(t *testing.T) {
	base, _, service := buildDecoratorFixture(t)

	created, err := service.Create(nil, dmodel.DynamicFields{"template_id": "t1", "template_name": "Injected"})
	require.NoError(t, err)
	require.Positive(t, created.ClientErrors.Count())
	assert.Contains(t, created.ClientErrors.ToError().Error(), `Field "template_name" is computed and cannot be written`)
	assert.Equal(t, 0, base.calls, "the write must be rejected before it reaches the base service")

	updated, err := service.Update(nil, dmodel.DynamicFields{"id": "v1", "template_name": "Injected"})
	require.NoError(t, err)
	assert.Positive(t, updated.ClientErrors.Count())

	ok, err := service.Update(nil, dmodel.DynamicFields{"id": "v1"})
	require.NoError(t, err)
	assert.Zero(t, ok.ClientErrors.Count(), "ordinary writes pass through")
}

// A search that names no fields still gets a narrow projection, the schema's
// default_search_fields, so the FK operand must be appended for it just as it is for an
// explicit projection.
func TestComputedServiceSearchDefaultProjectionCarriesOperands(t *testing.T) {
	base, recorder, service := buildDecoratorFixtureWithDefaults(t, []string{"template_name", "id"})
	base.searchPage = []dmodel.DynamicFields{{"id": "v1", "template_id": "t1"}}
	recorder.sourceRows = []dmodel.DynamicFields{{"id": "t1", "name": "Widget"}}

	result, err := service.Search(nil, dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Equal(t, []string{"template_name", "id", "template_id"}, base.gotParams["fields"])
	assert.Equal(t, 1, recorder.calls)
	assert.Equal(t, "Widget", result.Data.Items[0]["template_name"])
}

func TestComputedServiceSearchNamedViewSkipsComputed(t *testing.T) {
	base, recorder, service := buildDecoratorFixtureWithDefaults(t, []string{"template_name", "id"})
	base.searchPage = []dmodel.DynamicFields{{"id": "v1", "template_id": "t1"}}

	_, err := service.Search(nil, dmodel.DynamicFields{"search_name": "my_view"})

	require.NoError(t, err)
	assert.Zero(t, recorder.calls, "an id-only view must not trigger a related read")
}

func TestComputedServiceDefaultProjectionEvaluatesOnlyListedComputedFields(t *testing.T) {
	base, recorder, service := buildDecoratorFixtureWithDefaults(t, []string{"id", "template_id"})
	base.searchPage = []dmodel.DynamicFields{{"id": "v1", "template_id": "t1"}}

	result, err := service.Search(nil, dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Zero(t, recorder.calls)
	assert.NotContains(t, result.Data.Items[0], "template_name")
}

// Options pass through the decorator untouched, so a derived service's rules still reach the
// crud helper when the onion has wrapped the default.
func TestComputedServiceForwardsOptions(t *testing.T) {
	inner := newFakeDomainService(newPlainSchema())
	service := WithComputedFields(inner, nil, nil, nil)
	var calls []string

	_, err := service.Create(nil, dmodel.DynamicFields{"name": "x"}, CreateOptions{ValidateExtra: recording(&calls, "rule")})
	require.NoError(t, err)
	require.NotNil(t, inner.createOpts.ValidateExtra)
}
