package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestNewDomainServiceUsesConfiguredDefaultFields(t *testing.T) {
	service := NewDefaultDomainService(NewDomainServiceParam{
		Schema:        newPlainSchema(),
		DefaultFields: []string{"name"},
	}).(*DefaultDomainServiceImpl)

	assert.Equal(t, []string{"name"}, service.defaultFields)
}

func TestNewDomainServiceFallsBackToAllColumns(t *testing.T) {
	schema := newPlainSchema()
	service := NewDefaultDomainService(NewDomainServiceParam{Schema: schema}).(*DefaultDomainServiceImpl)

	assert.Equal(t, columnNames(schema), service.defaultFields)
	assert.Contains(t, service.defaultFields, "name")
}

func TestDomainServiceRefusesActionLeftOutOfCrudActions(t *testing.T) {
	service := NewDefaultDomainService(NewDomainServiceParam{
		Schema:      newPlainSchema(),
		CrudActions: []CrudAction{CrudActionGetById, CrudActionSearch},
	})

	result, err := service.Create(nil, dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err, "an unsupported action is a client error, not a Go error")
	assert.True(t, hasUnsupportedActionError(result.ClientErrors))
}

func TestParamsToSearchQueryDecodesComputedContext(t *testing.T) {
	query, err := paramsToSearchQuery(dmodel.DynamicFields{
		"fields":  []string{"id", "available_qty"},
		"context": map[string]any{"warehouse_id": "wh1", "company_id": "co1"},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"warehouse_id": "wh1", "company_id": "co1"}, query.Context)
}

func TestParamsToSearchQueryReportsAWrongTypedParamAsClientError(t *testing.T) {
	_, err := paramsToSearchQuery(dmodel.DynamicFields{"page": "two"})
	require.Error(t, err)

	cErrs, ok := clientErrorsForDecodeFailure(err)
	assert.True(t, ok, "a wrong-typed query parameter is the caller's mistake")
	assert.Positive(t, cErrs.Count())
}

func violation(key string) ValidateExtraFn {
	return func(_ corectx.Context, _ *DynamicEntity, _ *DynamicEntity, vErrs *ft.ClientErrors) error {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey(key), key))
		return nil
	}
}

func recording(calls *[]string, name string) ValidateExtraFn {
	return func(_ corectx.Context, _ *DynamicEntity, _ *DynamicEntity, _ *ft.ClientErrors) error {
		*calls = append(*calls, name)
		return nil
	}
}

func TestChainValidateExtraRunsInOrderAndSkipsNil(t *testing.T) {
	var calls []string
	chain := ChainValidateExtra(nil, recording(&calls, "first"), nil, recording(&calls, "second"))

	require.NoError(t, chain(nil, NewDynamicEntity(), nil, newClientErrors()))

	assert.Equal(t, []string{"first", "second"}, calls)
}

// A guard placed first keeps a later rule from running against input it already refused, which
// is the capture-then-chain behaviour inventory's create guard relied on.
func TestChainValidateExtraStopsAtTheFirstViolation(t *testing.T) {
	var calls []string
	chain := ChainValidateExtra(violation("guard"), recording(&calls, "rule"))

	vErrs := newClientErrors()
	require.NoError(t, chain(nil, NewDynamicEntity(), nil, vErrs))

	assert.Equal(t, 1, vErrs.Count())
	assert.Empty(t, calls, "the rule must not run once the guard refused the input")
}

func TestChainValidateExtraOfNothingIsNil(t *testing.T) {
	assert.Nil(t, ChainValidateExtra())
	assert.Nil(t, ChainValidateExtra(nil, nil))
}

func TestRejectArchivedOnCreateRefusesTheFlagOnlyWhenSent(t *testing.T) {
	guard := RejectArchivedOnCreate(newOrgSchema(), dmodel.DynamicFields{basemodel.FieldIsArchived: true})
	require.NotNil(t, guard)

	vErrs := newClientErrors()
	require.NoError(t, guard(nil, NewDynamicEntity(), nil, vErrs))
	assert.Equal(t, 1, vErrs.Count())
}

// corecrud injects the schema's default for is_archived before ValidateExtra runs, so a create
// the client sent without the field arrives at the hook carrying it. The guard therefore decides
// from the command, and a command that never named the field yields no hook at all — without
// this, every create of an archivable resource was refused.
func TestRejectArchivedOnCreateIgnoresASchemaInjectedDefault(t *testing.T) {
	assert.Nil(t, RejectArchivedOnCreate(newOrgSchema(), dmodel.DynamicFields{"name": "x"}))
}

func TestRejectArchivedOnCreateIsNilForASchemaWithoutTheField(t *testing.T) {
	sent := dmodel.DynamicFields{basemodel.FieldIsArchived: true}
	assert.Nil(t, RejectArchivedOnCreate(newPlainSchema(), sent))
	assert.Nil(t, RejectArchivedOnCreate(nil, sent))
}

// The onion's create guard runs ahead of whatever the derived service passes, on every Create.
func TestCreateGuardIsChainedAheadOfTheCallersHook(t *testing.T) {
	domSvc := newFakeDomainService(newOrgSchema())
	var calls []string
	newGuard := func(_ *dmodel.ModelSchema, _ CreateCommand) ValidateExtraFn {
		return recording(&calls, "guard")
	}
	guarded := withCreateGuard(domSvc, newGuard)

	_, err := guarded.Create(nil, dmodel.DynamicFields{}, CreateOptions{ValidateExtra: recording(&calls, "rule")})
	require.NoError(t, err)
	require.NotNil(t, domSvc.createOpts.ValidateExtra)
	require.NoError(t, domSvc.createOpts.ValidateExtra(nil, NewDynamicEntity(), nil, newClientErrors()))

	assert.Equal(t, []string{"guard", "rule"}, calls)
}

func TestWithCreateGuardIsANoOpWithoutAGuard(t *testing.T) {
	domSvc := newFakeDomainService(newPlainSchema())
	assert.Same(t, CrudDomainService(domSvc), withCreateGuard(domSvc, nil))
}
