package composable

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func newClientErrors() *ft.ClientErrors {
	cErrs := ft.ClientErrors{}
	return &cErrs
}

func TestApplicationServiceDeniesACallerWithoutPermission(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema())

	result, err := appSvc.Create(deniedContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err, "a denial is a client error, not a Go error")
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestApplicationServiceDelegatesForAnOwner(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema())

	result, err := appSvc.Create(ownerContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, "x", domSvc.seen["name"])
}

// GetSchema and ComputeField touch no stored row, so they never demand an org even on an
// org-bearing resource; the schema bootstrap of a client must not deadlock on the org it is
// about to learn about.
func TestSchemaAndComputeFieldAreNeverOrgScoped(t *testing.T) {
	_, appSvc := newAppService(newOrgSchema())

	schema, err := appSvc.GetSchema(memberContext(), dmodel.DynamicFields{})
	require.NoError(t, err)
	assert.Zero(t, schema.ClientErrors.Count())
	assert.True(t, schema.HasData)

	computed, err := appSvc.ComputeField(memberContext(), dmodel.DynamicFields{"field": "name"})
	require.NoError(t, err)
	assert.Zero(t, computed.ClientErrors.Count())
}

func TestApplicationServiceRefusesActionLeftOutOfCrudActions(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{
		CrudActions: []CrudAction{CrudActionGetById, CrudActionSearch},
	})

	result, err := appSvc.Create(ownerContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err, "an unsupported action is a client error, not a Go error")
	assert.True(t, hasUnsupportedActionError(result.ClientErrors))
	assert.Zero(t, domSvc.calls)

	search, err := appSvc.Search(ownerContext(), dmodel.DynamicFields{})
	require.NoError(t, err)
	assert.False(t, hasUnsupportedActionError(search.ClientErrors), "a listed action is not refused")
}

func TestResourceCodeIsTheSchemaName(t *testing.T) {
	_, appSvc := newAppService(newPlainSchema())
	assert.Equal(t, "cmp_plain_resource", appSvc.ResourceCode())
}

func hasUnsupportedActionError(cErrs ft.ClientErrors) bool {
	for _, item := range cErrs {
		if strings.HasSuffix(string(item.Key), "err_action_not_supported") {
			return true
		}
	}
	return false
}

// org_id is a declared required-for-create field, so a create that omits it must answer the same
// err_missing_required_field every other required field answers. Reporting err_org_id_required
// here would make one field explain itself differently from its neighbours.
func TestCreateDefersAMissingOrgIdToSchemaValidation(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgRequiredSchema())

	result, err := appSvc.Create(ownerContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count(), "the org resolver must not pre-empt schema validation")
	assert.Equal(t, 1, domSvc.calls, "the create reaches the domain service, which validates the schema")
}

// A create that names an org still resolves it, so a caller naming an org they do not belong to
// is refused before the record is written.
func TestCreateStillRefusesAForeignOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgRequiredSchema())

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{
		"name": "x", basemodel.FieldOrgId: "org_someone_else",
	})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}
