package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
)

func newTestResourceAccess(appSvc CrudApplicationService) *engineResourceAccess {
	return &engineResourceAccess{
		resolve:  func(string) (CrudApplicationService, error) { return appSvc, nil },
		schemaOf: func(string) *dmodel.ModelSchema { return appSvc.Schema() },
	}
}

func orgQuery(id string) resourceaccess.ReadAsCallerQuery {
	orgId := model.Id("org_mine")
	return resourceaccess.ReadAsCallerQuery{ResourceCode: "cmp_org_resource", ResourceId: model.Id(id), OrgId: &orgId}
}

func TestReadAsCallerRunsTheTargetGetById(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())
	domSvc.repo.(*stubRepository).found = true

	result, refused, err := newTestResourceAccess(appSvc).ReadAsCaller(memberContext(), orgQuery("r1"))

	require.NoError(t, err)
	require.Empty(t, refused)
	assert.True(t, result.Exists)
	assert.Equal(t, "r1", domSvc.seen[basemodel.FieldId])
	assert.Equal(t, columnNames(newOrgSchema()), domSvc.seen[basemodel.FieldFields])
}

func TestReadAsCallerRefusesWithoutRead(t *testing.T) {
	_, appSvc := newAppService(newOrgSchema())

	result, refused, err := newTestResourceAccess(appSvc).ReadAsCaller(deniedContext(), orgQuery("r1"))

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.NotEmpty(t, refused)
}

func TestReadAsCallerFallsBackToResourceReadForADeletedRecord(t *testing.T) {
	_, appSvc := newAppService(newOrgSchema())

	result, refused, err := newTestResourceAccess(appSvc).ReadAsCaller(memberContext(), orgQuery("gone"))

	require.NoError(t, err)
	require.Empty(t, refused)
	assert.False(t, result.Exists)
}

func TestReadAsCallerAnswersAnUnknownResource(t *testing.T) {
	access := &engineResourceAccess{schemaOf: func(string) *dmodel.ModelSchema { return nil }}

	_, refused, err := access.ReadAsCaller(memberContext(), orgQuery("r1"))

	require.NoError(t, err)
	assert.NotEmpty(t, refused)
}

func TestGetByIdIsAuditedOnlyWhenOptedIn(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	_, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{AuditSink: recorder})
	_, quietSvc := newAppService(newPlainSchema())

	_, err := appSvc.GetById(ownerContext(), dmodel.DynamicFields{basemodel.FieldId: "r1"})
	require.NoError(t, err)
	_, err = quietSvc.GetById(ownerContext(), dmodel.DynamicFields{basemodel.FieldId: "r1"})
	require.NoError(t, err)

	require.Len(t, recorder.entries, 1)
	assert.Equal(t, audit.ActionGet, recorder.entries[0].Action)
	assert.Equal(t, "r1", *recorder.entries[0].ResourceId)
}
