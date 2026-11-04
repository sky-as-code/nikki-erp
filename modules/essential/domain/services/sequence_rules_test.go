package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

const testOrgId = "01KNV107H8P0V2CB1R9T9ZNMF1"

func incrementDefinition(overrides dmodel.DynamicFields) *models.SequenceDefinition {
	fields := dmodel.DynamicFields{
		models.SeqDefFieldId: "01KNV107H8P0V2CB1R9T9ZNMF2", models.SeqDefFieldConfigurationScope: "org",
		models.SeqDefFieldOrgId: testOrgId, models.SeqDefFieldAlgorithm: models.SeqAlgorithmIncrement,
		models.SeqDefFieldFormatOwner: models.SeqFormatOwnerSequence, models.SeqDefFieldPadding: int32(5),
		models.SeqDefFieldIncrement: int64(1), models.SeqDefFieldInitialValue: int64(1),
		models.SeqDefFieldResetPeriod: models.SeqResetNever, models.SeqDefFieldTimezone: "UTC",
	}
	for key, value := range overrides {
		fields[key] = value
	}
	return models.NewSequenceDefinitionFrom(fields)
}

func errorKeys(vErrs *ft.ClientErrors) []string {
	keys := make([]string, 0, vErrs.Count())
	for _, item := range *vErrs {
		keys = append(keys, item.Field+"|"+item.Key)
	}
	return keys
}

func runCreateNormalization(t *testing.T, fields dmodel.DynamicFields) (dmodel.DynamicFields, *ft.ClientErrors) {
	t.Helper()
	ctx := corectx.NewRequestContext(context.Background())
	vErrs := ft.NewClientErrors()
	hook := normalizeDefinitionCreate(func(corectx.Context) string { return "Asia/Ho_Chi_Minh" })
	out, err := hook(ctx, dynamicresource.NewDynamicEntityFrom(fields), vErrs)
	require.NoError(t, err)
	return out.GetFieldData(), vErrs
}

func TestNormalizeDefinitionCreate_ScopeTypeAndDefaults(t *testing.T) {
	fields, vErrs := runCreateNormalization(t, dmodel.DynamicFields{
		models.SeqDefFieldConfigurationScope: "org", models.SeqDefFieldOrgId: testOrgId,
		models.SeqDefFieldSubScope: "Inventory_OpType", models.SeqDefFieldScopeType: "forged",
		models.SeqDefFieldAlgorithmParams: map[string]any{"key": "forged"},
		models.SeqDefFieldFirstIssuedAt:   "2026-01-01T00:00:00Z",
	})
	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "inventory_optype", fields[models.SeqDefFieldSubScope])
	assert.Equal(t, "org.inventory_optype", fields[models.SeqDefFieldScopeType])
	assert.Nil(t, fields[models.SeqDefFieldAlgorithmParams], "increment keeps no params")
	assert.NotContains(t, fields, models.SeqDefFieldFirstIssuedAt)
	assert.Equal(t, "Asia/Ho_Chi_Minh", fields[models.SeqDefFieldTimezone])
}

func TestNormalizeDefinitionCreate_PseudorandomGetsKeyAndNoPadding(t *testing.T) {
	fields, _ := runCreateNormalization(t, dmodel.DynamicFields{
		models.SeqDefFieldConfigurationScope: "tenant", models.SeqDefFieldAlgorithm: models.SeqAlgorithmFf1Fpe,
		models.SeqDefFieldPadding: float64(5),
	})
	assert.Equal(t, int32(0), fields[models.SeqDefFieldPadding])
	assert.Equal(t, "tenant", fields[models.SeqDefFieldScopeType])
	key := models.JsonMapString(fields[models.SeqDefFieldAlgorithmParams], models.SeqAlgorithmParamKey)
	assert.Len(t, key, 64, "32 random bytes, hex encoded")
}

func TestNormalizeDefinitionCreate_OwnershipShape(t *testing.T) {
	_, vErrs := runCreateNormalization(t, dmodel.DynamicFields{
		models.SeqDefFieldConfigurationScope: "tenant", models.SeqDefFieldOrgId: testOrgId,
	})
	assert.Equal(t, []string{"org_id|" + ErrKeySeqInvalidInput}, errorKeys(vErrs))

	_, vErrs = runCreateNormalization(t, dmodel.DynamicFields{models.SeqDefFieldConfigurationScope: "org"})
	assert.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "org_id", (*vErrs)[0].Field)
}

func TestValidateDefinitionConfig(t *testing.T) {
	cases := []struct {
		name      string
		overrides dmodel.DynamicFields
		want      []string
	}{
		{"valid", nil, []string{}},
		{"increment needs padding", dmodel.DynamicFields{models.SeqDefFieldPadding: int32(0)},
			[]string{"padding|" + ErrKeySeqPadding}},
		{"pseudorandom forbids padding", dmodel.DynamicFields{
			models.SeqDefFieldAlgorithm: models.SeqAlgorithmFf1Fpe, models.SeqDefFieldPadding: int32(3)},
			[]string{"padding|" + ErrKeySeqPadding}},
		{"bad timezone", dmodel.DynamicFields{models.SeqDefFieldTimezone: "Mars/Olympus"},
			[]string{"timezone|" + ErrKeySeqInvalidTimezone}},
		{"unknown variable", dmodel.DynamicFields{models.SeqDefFieldPrefixTemplate: "SO/{WH}/"},
			[]string{"prefix_template|" + ErrKeySeqInvalidTemplate}},
		{"yearly without period marker", dmodel.DynamicFields{
			models.SeqDefFieldResetPeriod: models.SeqResetYearly, models.SeqDefFieldPrefixTemplate: "SO/"},
			[]string{"prefix_template|" + ErrKeySeqInvalidTemplate}},
		{"yearly with period marker", dmodel.DynamicFields{
			models.SeqDefFieldResetPeriod: models.SeqResetYearly, models.SeqDefFieldPrefixTemplate: "SO/{YYYY}/"},
			[]string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vErrs := ft.NewClientErrors()
			ValidateDefinitionConfig(incrementDefinition(tc.overrides), vErrs)
			assert.Equal(t, tc.want, errorKeys(vErrs))
		})
	}
}

func TestAssertConfigurationUnlocked(t *testing.T) {
	stored := incrementDefinition(dmodel.DynamicFields{models.SeqDefFieldPrefixTemplate: "SO"}).GetFieldData()
	vErrs := ft.NewClientErrors()
	AssertConfigurationUnlocked(nil, dmodel.DynamicFields{
		models.SeqDefFieldName: "Orders", models.SeqDefFieldDescription: "new",
		models.SeqDefFieldPrefixTemplate: "SO", models.SeqDefFieldPadding: int32(5), "etag": "1",
	}, stored, vErrs)
	assert.Zero(t, vErrs.Count(), "name, description and unchanged values pass")

	AssertConfigurationUnlocked(nil, dmodel.DynamicFields{models.SeqDefFieldPadding: int32(6)}, stored, vErrs)
	assert.Equal(t, []string{"padding|" + ErrKeySeqConfigurationLocked}, errorKeys(vErrs))
}

func TestBuildScopeKey(t *testing.T) {
	plain := incrementDefinition(nil)
	key, cErrs := BuildScopeKey(plain, "T1", nil)
	assert.Nil(t, cErrs)
	assert.Equal(t, "tenant:T1/org:"+testOrgId, key)

	key, _ = BuildScopeKey(plain, "", nil)
	assert.Equal(t, "org:"+testOrgId, key, "single-tenant binary has no tenant part")

	objectId := model.Id("01KNV107H8P0V2CB1R9T9ZNMF3")
	_, cErrs = BuildScopeKey(plain, "T1", &objectId)
	require.NotNil(t, cErrs, "a plain scope refuses an object id")

	custom := incrementDefinition(dmodel.DynamicFields{models.SeqDefFieldSubScope: "inventory_optype"})
	_, cErrs = BuildScopeKey(custom, "T1", nil)
	require.NotNil(t, cErrs, "a custom scope requires an object id")
	key, cErrs = BuildScopeKey(custom, "T1", &objectId)
	assert.Nil(t, cErrs)
	assert.Equal(t, "tenant:T1/org:"+testOrgId+"/inventory_optype:"+string(objectId), key)
}

func TestFormatNumber_Increment(t *testing.T) {
	def := incrementDefinition(dmodel.DynamicFields{
		models.SeqDefFieldPrefixTemplate: "SO/{YYYY}/", models.SeqDefFieldSuffixTemplate: "-X",
	})
	at := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	vars := seqalgo.TemplateVars{EffectiveAt: at, AcceptedAt: at, PeriodKey: seqalgo.PeriodKeyNone}
	formatted, cErrs, err := FormatNumber(def, &itSeq.IncrementedCounter{LastNumber: 42}, vars)
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Equal(t, "SO/2026/00042-X", formatted)

	consumer := incrementDefinition(dmodel.DynamicFields{
		models.SeqDefFieldFormatOwner: models.SeqFormatOwnerConsumer, models.SeqDefFieldPrefixTemplate: "SO",
	})
	formatted, _, _ = FormatNumber(consumer, &itSeq.IncrementedCounter{LastNumber: 7}, vars)
	assert.Equal(t, "00007", formatted, "a consumer gets the padded number only")
}

func TestFormatNumber_Pseudorandom(t *testing.T) {
	key, err := seqalgo.NewKeyHex()
	require.NoError(t, err)
	tweak, err := seqalgo.NewTweakHex()
	require.NoError(t, err)
	def := incrementDefinition(dmodel.DynamicFields{
		models.SeqDefFieldAlgorithm: models.SeqAlgorithmFf1Fpe, models.SeqDefFieldPadding: int32(0),
		models.SeqDefFieldPrefixTemplate: "CORE", models.SeqDefFieldAlgorithmParams: map[string]any{"key": key},
	})
	params, _ := json.Marshal(map[string]string{"tweak": tweak})
	formatted, cErrs, err := FormatNumber(def, &itSeq.IncrementedCounter{LastNumber: 1, AlgorithmParams: params},
		seqalgo.TemplateVars{AcceptedAt: time.Now(), EffectiveAt: time.Now()})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Len(t, formatted, len("CORE")+seqalgo.MinLength)
	assert.Equal(t, "CORE", formatted[:4])
}

func TestUpsertFields(t *testing.T) {
	sub := "inventory_optype"
	orgId := model.Id(testOrgId)
	fields := UpsertFields(itSeq.UpsertDefinitionCommand{
		ConfigurationScope: "tenant", OrgId: &orgId, OwnerModule: "inventory", Code: "inventory.transfer",
		SubScope: &sub, Names: model.LangJson{"en-US": "Transfers", "vi-VN": "Điều chuyển"},
	}, model.LanguageCodeViVn)
	assert.NotContains(t, fields, models.SeqDefFieldOrgId, "a tenant definition drops the org")
	assert.Equal(t, sub, fields[models.SeqDefFieldSubScope])
	assert.Equal(t, "Điều chuyển", fields[models.SeqDefFieldName], "the name is stored in the given language")
	assert.NotContains(t, fields, models.SeqDefFieldPadding, "unset options fall back to schema defaults")
	assert.Equal(t, "tenant.inventory_optype", ScopeTypeOf("tenant", "Inventory_OpType"))
}

func TestNewNextValueRequest_Validation(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	badId := "not-a-uuid"
	_, cErrs := newNextValueRequest(ctx, itSeq.NextValueCommand{IdempotencyId: &badId})
	require.NotNil(t, cErrs)
	assert.ElementsMatch(t, []string{
		"definition_id|" + ErrKeySeqInvalidInput, "idempotency_id|" + ErrKeySeqInvalidIdempotencyId,
	}, errorKeys(cErrs))

	id, err := corectx.NewIdempotencyId()
	require.NoError(t, err)
	req, cErrs := newNextValueRequest(ctx, itSeq.NextValueCommand{DefinitionId: "X", IdempotencyId: &id})
	assert.Nil(t, cErrs)
	assert.Equal(t, models.SeqDefaultOperationKey, req.operationKey)
	assert.Nil(t, req.effectiveAt, "no effective_at keeps the SERVER_DEFAULT fingerprint")
}

func TestAssertReplayable(t *testing.T) {
	future := model.WrapModelDateTime(time.Now().Add(time.Hour))
	past := model.WrapModelDateTime(time.Now().Add(-time.Hour))
	stored := dmodel.DynamicFields{models.SeqIssFieldRequestFingerprint: "abc", models.SeqIssFieldExpiresAt: future}
	assert.Nil(t, assertReplayable(stored, "abc"))
	assert.Equal(t, []string{"idempotency_id|" + ErrKeySeqIdempotencyConflict},
		errorKeys(assertReplayable(stored, "other")))
	stored[models.SeqIssFieldExpiresAt] = past
	assert.Equal(t, []string{"idempotency_id|" + ErrKeySeqIdempotencyExpired},
		errorKeys(assertReplayable(stored, "abc")))
}

// NextValue/lockOwned serve a tenant definition to every org, an org definition to its org alone.
func TestOwnedBy(t *testing.T) {
	orgA, orgB := model.Id("org-a"), model.Id("org-b")
	orgDef := dmodel.DynamicFields{"org_id": "org-a"}
	tenantDef := dmodel.DynamicFields{}
	assert.True(t, ownedBy(tenantDef, &orgA))
	assert.True(t, ownedBy(orgDef, &orgA))
	assert.False(t, ownedBy(orgDef, &orgB))
}
