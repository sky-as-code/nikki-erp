package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const edgeDestSchemaName = "cmp_edge_dest"

func newEdgeSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_edge_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50)).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name("dest_id").DataType(dmodel.FieldDataTypeString(0, 50))).
		EdgeTo(dmodel.Edge("dest").
			ManyToOne(edgeDestSchemaName, dmodel.DynamicFields{"dest_id": "id"}).
			OnDelete(dmodel.RelationCascadeSetNull)).
		Build()
}

// edgeDomainService answers every read with a record carrying the edge, as the database would.
type edgeDomainService struct {
	*fakeDomainService
}

func (this *edgeDomainService) item() dmodel.DynamicFields {
	return dmodel.DynamicFields{"id": "r1", "name": "x", "dest": dmodel.DynamicFields{"name": "d"}}
}

func (this *edgeDomainService) GetById(_ corectx.Context, query GetByIdQuery) (*GetOneResult, error) {
	this.record(query)
	data := dyn.SingleResultData[dmodel.DynamicFields]{Item: this.item()}
	data.Meta.DesiredFields = requestedFields(query)
	return &GetOneResult{Data: data, HasData: true}, nil
}

func (this *edgeDomainService) Search(
	_ corectx.Context, query SearchQuery, _ ...corecrud.ServiceSearchOptions,
) (*SearchResult, error) {
	this.record(query)
	data := dyn.PagedResultData[dmodel.DynamicFields]{Items: []dmodel.DynamicFields{this.item()}}
	data.DesiredFields = requestedFields(query)
	return &SearchResult{Data: data, HasData: true}, nil
}

func newEdgeAppService() (*edgeDomainService, CrudApplicationService) {
	tenant := requestguard.ResourceScopeTenant
	domSvc := &edgeDomainService{newFakeDomainService(newEdgeSchema())}
	return domSvc, NewDefaultApplicationService(NewAppServiceParam{DomainService: domSvc, PermissionScope: &tenant})
}

func TestGetByIdMasksAnEdgeWhoseDestinationIsUnreadable(t *testing.T) {
	domSvc, appSvc := newEdgeAppService()
	ctx := grantedContext("read:cmp_edge_resource:tenant")

	result, err := appSvc.GetById(ctx, dmodel.DynamicFields{"id": "r1", "fields": []string{"name", "dest.name"}})
	require.NoError(t, err)
	require.Zero(t, result.ClientErrors.Count(), "projection never answers 403")

	assert.Equal(t, []string{"name"}, domSvc.seen["fields"], "the edge is never fetched")
	assert.Nil(t, result.Data.Item["dest"])
	assert.Contains(t, result.Data.Item, "dest")
	assert.Equal(t, []string{"dest.name"}, result.Data.Meta.MaskedFields)
	assert.Contains(t, result.Data.Meta.DesiredFields, "dest.name")
}

func TestSearchMasksABareEdgeAtTopLevel(t *testing.T) {
	_, appSvc := newEdgeAppService()
	ctx := grantedContext("read:cmp_edge_resource:tenant", "read:cmp_edge_dest.name:tenant")

	result, err := appSvc.Search(ctx, dmodel.DynamicFields{"fields": []string{"dest"}})
	require.NoError(t, err)
	require.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, []string{"dest"}, result.Data.MaskedFields, "a bare edge needs the whole destination")
	assert.Nil(t, result.Data.Items[0]["dest"])
}

func TestEdgeLeafReadableThroughADestinationFieldGrant(t *testing.T) {
	domSvc, appSvc := newEdgeAppService()
	ctx := grantedContext("read:cmp_edge_resource:tenant", "read:cmp_edge_dest.name:tenant")

	result, err := appSvc.Search(ctx, dmodel.DynamicFields{"fields": []string{"dest.name", "dest.code"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"dest.name"}, domSvc.seen["fields"])
	assert.Equal(t, []string{"dest.code"}, result.Data.MaskedFields)
	assert.NotNil(t, result.Data.Items[0]["dest"], "a readable leaf keeps the edge")
}

func TestEdgeReadableWithTheDestinationGrant(t *testing.T) {
	domSvc, appSvc := newEdgeAppService()
	ctx := grantedContext("read:cmp_edge_resource:tenant", "read:cmp_edge_dest:tenant")

	result, err := appSvc.GetById(ctx, dmodel.DynamicFields{"id": "r1", "fields": []string{"dest"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"dest"}, domSvc.seen["fields"])
	assert.Empty(t, result.Data.Meta.MaskedFields)
}

func TestEdgeMaskedWhenTheSourceEdgeFieldIsUnreadable(t *testing.T) {
	domSvc, appSvc := newEdgeAppService()
	ctx := grantedContext("read:cmp_edge_resource.name:tenant", "read:cmp_edge_dest:tenant")

	result, err := appSvc.GetById(ctx, dmodel.DynamicFields{"id": "r1", "fields": []string{"dest.name"}})
	require.NoError(t, err)
	require.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, []string{"id"}, domSvc.seen["fields"], "nothing left falls back to the key, not the defaults")
	assert.Contains(t, result.Data.Meta.MaskedFields, "dest.name")
	assert.Nil(t, result.Data.Item["dest"])
}

func TestGetSchemaNeedsNoPermission(t *testing.T) {
	_, appSvc := newEdgeAppService()
	result, err := appSvc.GetSchema(grantedContext(), dmodel.DynamicFields{})
	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.True(t, result.HasData)
}
