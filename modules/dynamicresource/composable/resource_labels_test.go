package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
)

// labelSearchApp answers Search with canned rows and remembers the queries it was asked.
type labelSearchApp struct {
	CrudApplicationService
	schema  *dmodel.ModelSchema
	items   []dmodel.DynamicFields
	refused bool
	queries []SearchQuery
}

func (this *labelSearchApp) Schema() *dmodel.ModelSchema { return this.schema }

func (this *labelSearchApp) Search(_ corectx.Context, query SearchQuery) (*SearchResult, error) {
	this.queries = append(this.queries, query)
	if this.refused {
		return &SearchResult{ClientErrors: ft.ClientErrors{*ft.NewAnonymousNotFoundError()}}, nil
	}
	return &SearchResult{Data: dyn.PagedResultData[dmodel.DynamicFields]{Items: this.items}, HasData: true}, nil
}

func newLabelledSchema(labelField string) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cmp_labelled_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200)))
	if labelField != "" {
		builder = builder.RecordLabelField(labelField)
	}
	return builder.Build()
}

func labelAccessOver(app *labelSearchApp) *engineResourceAccess {
	return &engineResourceAccess{
		resolve:  func(string) (CrudApplicationService, error) { return app, nil },
		schemaOf: func(string) *dmodel.ModelSchema { return app.schema },
	}
}

func TestLabelsAsCallerReadsTheRecordLabelInOneSearch(t *testing.T) {
	app := &labelSearchApp{schema: newLabelledSchema("name"), items: []dmodel.DynamicFields{
		{"id": "r1", "name": "Role one"},
		{"id": "r2", "name": map[string]any{"vi-VN": "Vai trò hai", "en-US": "Role two"}},
		{"id": "r3", "name": nil},
	}}
	orgId := model.Id("org_1")

	labels, err := labelAccessOver(app).LabelsAsCaller(ownerContext(), resourceaccess.LabelsQuery{
		ResourceCode: "cmp_labelled_resource", Ids: []model.Id{"r1", "r2", "r3", "gone"}, OrgId: &orgId,
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"r1": "Role one", "r2": "Role two"}, labels)
	require.Len(t, app.queries, 1, "one search for the whole id set")
	query := app.queries[0]
	assert.Equal(t, []string{basemodel.FieldId, "name"}, query[fieldNameFields])
	assert.Equal(t, true, query[basemodel.FieldIncludeArchived])
	assert.Equal(t, "org_1", query[basemodel.FieldOrgId])
	assert.Equal(t, map[string]any{"if": []any{"id", "in", "r1", "r2", "r3", "gone"}}, query[fieldNameGraph])
}

func TestLabelsAsCallerLeavesARefusedReadUnlabelled(t *testing.T) {
	app := &labelSearchApp{schema: newLabelledSchema("name"), refused: true}

	labels, err := labelAccessOver(app).LabelsAsCaller(ownerContext(), resourceaccess.LabelsQuery{
		ResourceCode: "cmp_labelled_resource", Ids: []model.Id{"r1"},
	})

	require.NoError(t, err)
	assert.Empty(t, labels)
}

func TestLabelsAsCallerNeedsARecordLabelField(t *testing.T) {
	app := &labelSearchApp{schema: newLabelledSchema("")}

	labels, err := labelAccessOver(app).LabelsAsCaller(ownerContext(), resourceaccess.LabelsQuery{
		ResourceCode: "cmp_labelled_resource", Ids: []model.Id{"r1"},
	})

	require.NoError(t, err)
	assert.Empty(t, labels)
	assert.Empty(t, app.queries, "no search without a label field to read")
}

func TestLabelTextPrefersTheCallerLanguage(t *testing.T) {
	vi := model.LanguageCode("vi-VN")
	texts := model.LangJson{"en-US": "Role", "vi-VN": "Vai trò"}

	assert.Equal(t, "Vai trò", labelText(texts, &vi))
	assert.Equal(t, "Role", labelText(texts, nil))
	assert.Equal(t, "Rôle", labelText(map[string]any{"fr-FR": "Rôle"}, nil))
}
