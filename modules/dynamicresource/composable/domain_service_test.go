package composable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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

func newLangJsonSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_langjson_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("label").DataType(dmodel.FieldDataTypeLangJson(0, 100))).
		Field(dmodel.DefineField().Name("note").DataType(dmodel.FieldDataTypeLangJson(0, 100))).
		Build()
}

// withEnabledLanguages installs a process-wide resolver for one test; nothing else installs one in
// this package's tests, so the cleanup restores nil.
func withEnabledLanguages(t *testing.T, fn dyn.EnabledLanguagesResolver) {
	t.Helper()
	dyn.SetEnabledLanguagesResolver(fn)
	t.Cleanup(func() { dyn.SetEnabledLanguagesResolver(nil) })
}

func enabledLanguages(codes ...model.LanguageCode) dyn.EnabledLanguagesResolver {
	return func(corectx.Context) ([]model.LanguageCode, error) { return codes, nil }
}

func runLangJsonHook(t *testing.T, hook BeforeValidationFn, data dmodel.DynamicFields) dmodel.DynamicFields {
	t.Helper()
	require.NotNil(t, hook)
	result, err := hook(newTestRequestCtx(), NewDynamicEntityFrom(data), newClientErrors())
	require.NoError(t, err)
	return result.GetFieldData()
}

func newTestRequestCtx() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

func TestLangJsonNormalizerStripsAndFills(t *testing.T) {
	withEnabledLanguages(t, enabledLanguages("vi-VN", "en-US"))
	hook := newLangJsonNormalizer(newLangJsonSchema())

	got := runLangJsonHook(t, hook, dmodel.DynamicFields{
		"label": map[string]any{"en-US": "Hi", "fr-FR": "Salut"},
	})

	assert.Equal(t, model.LangJson{"en-US": "Hi", "vi-VN": "Hi"}, got["label"])
	_, hasNote := got["note"]
	assert.False(t, hasNote, "a field the caller did not send stays absent")
}

func TestLangJsonNormalizerLeavesRefAndEmptyAlone(t *testing.T) {
	withEnabledLanguages(t, enabledLanguages("en-US", "vi-VN"))
	hook := newLangJsonNormalizer(newLangJsonSchema())

	got := runLangJsonHook(t, hook, dmodel.DynamicFields{
		"label": model.LangJson{"$ref": "a.b"},
		"note":  model.LangJson{},
	})

	assert.Equal(t, model.LangJson{"$ref": "a.b"}, got["label"])
	assert.Equal(t, model.LangJson{}, got["note"])
}

func TestLangJsonNormalizerWithoutResolverFillsPresentKeysOnly(t *testing.T) {
	withEnabledLanguages(t, nil)
	hook := newLangJsonNormalizer(newLangJsonSchema())

	got := runLangJsonHook(t, hook, dmodel.DynamicFields{"label": model.LangJson{"en-US": "", "fr-FR": "Salut"}})

	assert.Equal(t, model.LangJson{"en-US": "Salut", "fr-FR": "Salut"}, got["label"])
}

func TestLangJsonNormalizerLeavesNonStringEntriesToTheValidator(t *testing.T) {
	withEnabledLanguages(t, enabledLanguages("en-US", "vi-VN"))
	hook := newLangJsonNormalizer(newLangJsonSchema())
	raw := map[string]any{"en-US": "Hi", "vi-VN": 42}

	got := runLangJsonHook(t, hook, dmodel.DynamicFields{"label": raw})

	assert.Equal(t, raw, got["label"])
}

func TestLangJsonNormalizerIsNilWithoutLangJsonFields(t *testing.T) {
	assert.Nil(t, newLangJsonNormalizer(newPlainSchema()))
}

func TestLangJsonNormalizationRunsBeforeTheCallerHook(t *testing.T) {
	withEnabledLanguages(t, enabledLanguages("en-US", "vi-VN"))
	service := NewDefaultDomainService(NewDomainServiceParam{Schema: newLangJsonSchema()}).(*DefaultDomainServiceImpl)
	var seen any
	callerHook := func(_ corectx.Context, entity *DynamicEntity, _ *ft.ClientErrors) (*DynamicEntity, error) {
		seen = entity.GetFieldData()["label"]
		return entity, nil
	}

	runLangJsonHook(t, service.withLangJsonNormalization(callerHook), dmodel.DynamicFields{
		"label": model.LangJson{"vi-VN": "Chào"},
	})

	assert.Equal(t, model.LangJson{"en-US": "Chào", "vi-VN": "Chào"}, seen)
}

// langJsonRepository answers only what corecrud reads before the BeforeValidation hook runs.
type langJsonRepository struct {
	CrudRepository
	base langJsonBaseRepository
}

type langJsonBaseRepository struct {
	dyn.BaseDynamicRepository
	schema *dmodel.ModelSchema
}

func (this langJsonBaseRepository) Schema() *dmodel.ModelSchema { return this.schema }

func (this *langJsonRepository) GetBaseRepo() dyn.BaseDynamicRepository { return this.base }

func (this *langJsonRepository) Schema() *dmodel.ModelSchema { return this.base.schema }

func TestCreateAndUpdateFailWhenEnabledLanguagesCannotBeRead(t *testing.T) {
	withEnabledLanguages(t, func(corectx.Context) ([]model.LanguageCode, error) {
		return nil, errors.New("languages unavailable")
	})
	schema := newLangJsonSchema()
	service := NewDefaultDomainService(NewDomainServiceParam{
		Schema:     schema,
		Repository: &langJsonRepository{base: langJsonBaseRepository{schema: schema}},
	})
	cmd := dmodel.DynamicFields{"id": "r1", "label": model.LangJson{"en-US": "Hi"}}

	_, createErr := service.Create(newTestRequestCtx(), cmd)
	_, updateErr := service.Update(newTestRequestCtx(), cmd)

	require.Error(t, createErr)
	assert.Contains(t, createErr.Error(), "languages unavailable")
	require.Error(t, updateErr)
	assert.Contains(t, updateErr.Error(), "languages unavailable")
}

func TestChainBeforeValidationStopsAtAViolation(t *testing.T) {
	var calls []string
	record := func(name string, violate bool) BeforeValidationFn {
		return func(_ corectx.Context, entity *DynamicEntity, vErrs *ft.ClientErrors) (*DynamicEntity, error) {
			calls = append(calls, name)
			if violate {
				vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey(name), name))
			}
			return entity, nil
		}
	}
	chain := ChainBeforeValidation(nil, record("first", true), record("second", false))

	_, err := chain(nil, NewDynamicEntity(), newClientErrors())

	require.NoError(t, err)
	assert.Equal(t, []string{"first"}, calls)
	assert.Nil(t, ChainBeforeValidation(nil, nil))
}
