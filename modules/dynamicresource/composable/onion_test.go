package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// fakeBaseRepo stands in for the SQL-backed base repository: nothing here reaches a database.
type fakeBaseRepo struct {
	dyn.BaseDynamicRepository
	schema *dmodel.ModelSchema
}

func (this *fakeBaseRepo) Schema() *dmodel.ModelSchema { return this.schema }

func testBuildParam() BuildParam {
	return BuildParam{
		NewBaseRepoFn: func(param dyn.NewBaseRepoParam) dyn.BaseDynamicRepository {
			return &fakeBaseRepo{schema: param.Schema}
		},
	}
}

const onionTestSchema = "cmp_onion_resource"

func registerOnionTestSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	return dmodel.GetOrRegisterSchema(onionTestSchema, func() *dmodel.ModelSchemaBuilder {
		return dmodel.DefineModel(onionTestSchema).
			Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
			Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
			DefaultSearchFields("name")
	})
}

type derivedRepo struct{ CrudRepository }
type derivedDomSvc struct{ CrudDomainService }
type derivedAppSvc struct{ CrudApplicationService }

func TestBuildHandsEachDefaultToTheModulesFactory(t *testing.T) {
	registerOnionTestSchema(t)
	var gotRepo CrudRepository
	var gotDomSvc CrudDomainService
	var gotAppSvc CrudApplicationService

	onion, err := (&DynamicResourceEngineOnionImpl{
		SchemaName: onionTestSchema,
		NewRepositoryFn: func(base CrudRepository) CrudRepository {
			gotRepo = base
			return &derivedRepo{base}
		},
		NewDomainServiceFn: func(base CrudDomainService) CrudDomainService {
			gotDomSvc = base
			return &derivedDomSvc{base}
		},
		NewAppServiceFn: func(base CrudApplicationService) CrudApplicationService {
			gotAppSvc = base
			return &derivedAppSvc{base}
		},
	}).Build(testBuildParam())
	require.NoError(t, err)

	assert.Equal(t, onionTestSchema, onion.ResourceName())
	assert.IsType(t, &DefaultCrudRepositoryImpl{}, gotRepo)
	assert.IsType(t, &computedFieldService{}, gotDomSvc, "the domain default is handed over computed-wrapped")
	assert.IsType(t, &DefaultApplicationServiceImpl{}, gotAppSvc)

	assert.IsType(t, &derivedRepo{}, onion.Repository())
	assert.IsType(t, &derivedDomSvc{}, onion.DomainService())
	assert.IsType(t, &derivedAppSvc{}, onion.ApplicationService())
	assert.Same(t, onion.DomainService(), onion.ApplicationService().DomainService(),
		"the application default delegates to the module's derived domain service")
}

func TestBuildWithoutFactoriesUsesTheDefaults(t *testing.T) {
	registerOnionTestSchema(t)

	onion, err := (&DynamicResourceEngineOnionImpl{SchemaName: onionTestSchema}).Build(testBuildParam())
	require.NoError(t, err)

	assert.IsType(t, &DefaultCrudRepositoryImpl{}, onion.Repository())
	assert.IsType(t, &DefaultApplicationServiceImpl{}, onion.ApplicationService())
	assert.Equal(t, onionTestSchema, onion.Schema().Name())
}

func TestBuildIndexesTheResourceAsAComputedSource(t *testing.T) {
	registerOnionTestSchema(t)
	_, err := (&DynamicResourceEngineOnionImpl{SchemaName: onionTestSchema}).Build(testBuildParam())
	require.NoError(t, err)

	repo, ok := LookupSourceRepository(onionTestSchema)
	assert.True(t, ok)
	assert.NotNil(t, repo)
	assert.Contains(t, BuiltSchemaNames(), onionTestSchema)
}

func TestBuildInstallsTheCreateGuardWhenAsked(t *testing.T) {
	registerOnionTestSchema(t)
	var gotDomSvc CrudDomainService

	_, err := (&DynamicResourceEngineOnionImpl{
		SchemaName:             onionTestSchema,
		RejectArchivedOnCreate: true,
		NewDomainServiceFn: func(base CrudDomainService) CrudDomainService {
			gotDomSvc = base
			return base
		},
	}).Build(testBuildParam())
	require.NoError(t, err)

	// The wrapper is installed whenever the onion asks for the guard; the guard itself decides
	// per call, from the command, and resolves to nil for a schema with no is_archived field.
	assert.IsType(t, &createGuardService{}, gotDomSvc)
	assert.Nil(t, RejectArchivedOnCreate(dmodel.GetSchema(onionTestSchema), CreateCommand{"name": "x"}))
}

func TestBuildRefusesAnUnknownSchema(t *testing.T) {
	_, err := (&DynamicResourceEngineOnionImpl{SchemaName: "cmp_no_such_schema"}).Build(testBuildParam())
	require.Error(t, err)

	assert.Panics(t, func() {
		MustBuild(&DynamicResourceEngineOnionImpl{SchemaName: "cmp_no_such_schema"}, testBuildParam())
	})
}

func TestBuildRefusesAnEmptySchemaName(t *testing.T) {
	_, err := (&DynamicResourceEngineOnionImpl{}).Build(testBuildParam())
	require.Error(t, err)
}

func TestBuildRefusesToBuildTwice(t *testing.T) {
	registerOnionTestSchema(t)
	impl := &DynamicResourceEngineOnionImpl{SchemaName: onionTestSchema}
	_, err := impl.Build(testBuildParam())
	require.NoError(t, err)

	_, err = impl.Build(testBuildParam())
	require.Error(t, err)
}

func TestEngineDependencyName(t *testing.T) {
	assert.Equal(t, "dynengine_essential_uom", EngineDependencyName("essential_uom"))
}

func TestActionTypeHttpMethod(t *testing.T) {
	assert.Equal(t, "POST", ActionTypeCreate.HttpMethod())
	assert.Equal(t, "POST", ActionTypeGeneric.HttpMethod())
	assert.Equal(t, "POST", ActionTypeUpload.HttpMethod())
	assert.Equal(t, "GET", ActionTypeRead.HttpMethod())
	assert.Equal(t, "PATCH", ActionTypeUpdatePatch.HttpMethod())
	assert.Equal(t, "PUT", ActionTypeUpdateReplace.HttpMethod())
	assert.Equal(t, "DELETE", ActionTypeDelete.HttpMethod())
	assert.Empty(t, ActionType("bogus").HttpMethod())
	assert.False(t, ActionTypeUpload.HasRequestBody())
	assert.True(t, ActionTypeGeneric.HasRequestBody())
}
