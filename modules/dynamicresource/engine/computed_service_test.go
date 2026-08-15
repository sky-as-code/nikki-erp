package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// The decorator generalizes what modules used to hand-roll for template_*-style fields, so these
// tests pin the behaviors those hand-rolled overrides guaranteed: one batched source query per
// page, the FK forced into an explicit projection, distinct keys fetched once, a missing source
// row leaving fields absent, and untouched requests passing straight through.

// fakeBaseService is the wrapped resource service: it records the params it receives and returns
// a canned page.
type fakeBaseService struct {
	it.DynamicResourceService
	schema     *dmodel.ModelSchema
	gotParams  dmodel.DynamicFields
	searchPage []dmodel.DynamicFields
	single     dmodel.DynamicFields
	calls      int
}

func (this *fakeBaseService) Schema() *dmodel.ModelSchema { return this.schema }

func (this *fakeBaseService) Search(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.calls++
	this.gotParams = params
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.searchPage},
		HasData: true,
	}, nil
}

func (this *fakeBaseService) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	this.calls++
	this.gotParams = params
	return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{
		Data:    dyn.SingleResultData[dmodel.DynamicFields]{Item: this.single},
		HasData: true,
	}, nil
}

func (this *fakeBaseService) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.calls++
	return &dyn.OpResult[dmodel.DynamicFields]{Data: params, HasData: true}, nil
}

func (this *fakeBaseService) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls++
	return &dyn.OpResult[dyn.MutateResultData]{HasData: true}, nil
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
		ctx corectx.Context, schemaName string, keyColumn string, keys []any, fields []string,
	) ([]dmodel.DynamicFields, error) {
		this.calls++
		this.gotSchema, this.gotKeys, this.gotFields = schemaName, keys, fields
		return this.sourceRows, nil
	}
}

func buildDecoratorFixture(t *testing.T) (*fakeBaseService, *sourceSearchRecorder, it.DynamicResourceService) {
	t.Helper()
	source := dmodel.DefineModel("cfsvc_template").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
	owner := dmodel.DefineModel("cfsvc_variant").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name"))).
		EdgeTo(dmodel.Edge("template").ManyToOne("cfsvc_template", dmodel.DynamicFields{"template_id": "id"})).
		Build()

	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(source))
	require.NoError(t, reg.Register(owner))
	require.NoError(t, reg.FinalizeRelations())

	base := &fakeBaseService{schema: owner}
	recorder := &sourceSearchRecorder{}
	return base, recorder, WithComputedFields(base, recorder.fn())
}

func TestComputedService_SearchBatchesOneSourceQuery(t *testing.T) {
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

	result, err := service.Search(nil, dmodel.DynamicFields{
		"fields": []string{"id", "template_name"},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, base.calls, "the page itself costs one query")
	assert.Equal(t, 1, recorder.calls, "the source read costs exactly one more, however large the page")
	assert.Equal(t, "cfsvc_template", recorder.gotSchema)
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

func TestComputedService_ForcesFkIntoExplicitProjection(t *testing.T) {
	base, _, service := buildDecoratorFixture(t)
	base.searchPage = []dmodel.DynamicFields{}

	_, err := service.Search(nil, dmodel.DynamicFields{
		"fields": []string{"id", "template_name"},
	})
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"id", "template_name", "template_id"},
		base.gotParams["fields"],
		"the FK join key must ride along, or the rows cannot be matched back to their sources")
}

func TestComputedService_UntouchedWhenNothingComputedRequested(t *testing.T) {
	base, recorder, service := buildDecoratorFixture(t)
	base.searchPage = []dmodel.DynamicFields{{"id": "v1", "template_id": "t1"}}

	_, err := service.Search(nil, dmodel.DynamicFields{
		"fields": []string{"id", "template_id"},
	})
	require.NoError(t, err)

	assert.Equal(t, 0, recorder.calls, "no computed field requested, no source read")
	assert.Equal(t, []string{"id", "template_id"}, base.gotParams["fields"],
		"the projection must pass through unmodified")
}

func TestComputedService_GetByIdFillsSingleRecord(t *testing.T) {
	base, recorder, service := buildDecoratorFixture(t)
	base.single = dmodel.DynamicFields{"id": "v1", "template_id": "t1"}
	recorder.sourceRows = []dmodel.DynamicFields{{"id": "t1", "name": "Widget"}}

	result, err := service.GetById(nil, dmodel.DynamicFields{
		"fields": []string{"id", "template_name"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Widget", result.Data.Item["template_name"])
}

func TestComputedService_WritesToComputedFieldRejected(t *testing.T) {
	base, _, service := buildDecoratorFixture(t)

	created, err := service.Create(nil, dmodel.DynamicFields{
		"template_id":   "t1",
		"template_name": "Injected",
	})
	require.NoError(t, err)
	require.Greater(t, created.ClientErrors.Count(), 0)
	assert.Contains(t, created.ClientErrors.ToError().Error(),
		`Field "template_name" is computed and cannot be written`)
	assert.Equal(t, 0, base.calls, "the write must be rejected before it reaches the base service")

	updated, err := service.Update(nil, dmodel.DynamicFields{
		"id":            "v1",
		"template_name": "Injected",
	})
	require.NoError(t, err)
	assert.Greater(t, updated.ClientErrors.Count(), 0)

	ok, err := service.Update(nil, dmodel.DynamicFields{"id": "v1"})
	require.NoError(t, err)
	assert.Equal(t, 0, ok.ClientErrors.Count(), "ordinary writes pass through")
}
