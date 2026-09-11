package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// newDedupSchema declares the two deduplication fields, which switches bulk create to upsert.
func newDedupSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_dedup_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name(basemodel.FieldOrgId).DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name(FieldSourceSystem).DataType(dmodel.FieldDataTypeString(0, 50)).
			Default("manual")).
		Field(dmodel.DefineField().Name(FieldExternalId).DataType(dmodel.FieldDataTypeString(0, 100))).
		Field(dmodel.DefineField().Name(basemodel.FieldEtag).DataType(dmodel.FieldDataTypeString(0, 50))).
		Build()
}

func bulkRows(names ...string) []BulkRow {
	rows := make([]BulkRow, 0, len(names))
	for i, name := range names {
		rows = append(rows, BulkRow{Number: i + 1, Fields: dmodel.DynamicFields{"name": name}})
	}
	return rows
}

func TestParamsToBulkRowsRejectsShapeProblems(t *testing.T) {
	cases := map[string]BulkCreateCommand{
		"missing":    {},
		"not a list": {"items": "x"},
		"not object": {"items": []any{"x"}},
	}
	for name, cmd := range cases {
		t.Run(name, func(t *testing.T) {
			rows, cErrs := paramsToBulkRows(cmd)
			assert.Nil(t, rows)
			require.NotNil(t, cErrs)
			assert.Equal(t, "items", (*cErrs)[0].Field)
		})
	}
}

func TestParamsToBulkRowsNumbersRowsFromOne(t *testing.T) {
	rows, cErrs := paramsToBulkRows(BulkCreateCommand{"items": []any{
		map[string]any{"name": "a"}, map[string]any{"name": "b"},
	}})

	require.Nil(t, cErrs)
	assert.Equal(t, []BulkRow{
		{Number: 1, Fields: dmodel.DynamicFields{"name": "a"}},
		{Number: 2, Fields: dmodel.DynamicFields{"name": "b"}},
	}, rows)
}

func TestRunBulkCreateSkipsRejectedRowsAndCommitsTheRest(t *testing.T) {
	domSvc := newFakeDomainService(newPlainSchema())
	domSvc.rejectName = "bad"

	result, err := RunBulkCreate(ownerContext(), domSvc, bulkRows("ok1", "bad", "ok2"), nil)

	require.NoError(t, err)
	require.True(t, result.HasData)
	assert.Equal(t, 2, result.Data.CreatedCount)
	assert.Equal(t, 0, result.Data.UpdatedCount)
	assert.Equal(t, 2, result.Data.AffectedCount)
	assert.Equal(t, 3, result.Data.TotalRows)
	require.Len(t, result.Data.Errors, 1)
	assert.Equal(t, RowError{
		Row: 2, Field: "name", Code: ErrRowValidation,
		Params: result.Data.Errors[0].Params,
	}, result.Data.Errors[0])
	assert.True(t, domSvc.repo.(*stubRepository).tranx.committed)
}

func TestRunBulkCreateEchoesCallerRejectedRowsInOrder(t *testing.T) {
	domSvc := newFakeDomainService(newPlainSchema())
	rejected := []RowError{{Row: 5, Code: "err_import_cell_invalid"}}

	result, err := RunBulkCreate(ownerContext(), domSvc, []BulkRow{{Number: 6, Fields: dmodel.DynamicFields{"name": "x"}}}, rejected)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Data.TotalRows)
	assert.Equal(t, 1, result.Data.ErrorCount)
	assert.Equal(t, 5, result.Data.Errors[0].Row)
}

func TestRunBulkCreateUpsertsByExternalKey(t *testing.T) {
	domSvc := newFakeDomainService(newDedupSchema())
	repo := domSvc.repo.(*stubRepository)
	repo.searchItems = []dmodel.DynamicFields{
		{"id": "stored_1", FieldSourceSystem: "manual", FieldExternalId: "E1", basemodel.FieldOrgId: "org_mine", basemodel.FieldEtag: "v7"},
	}
	rows := []BulkRow{
		{Number: 1, Fields: dmodel.DynamicFields{"name": "a", FieldExternalId: "E1", basemodel.FieldOrgId: "org_mine"}},
		{Number: 2, Fields: dmodel.DynamicFields{"name": "b", FieldExternalId: "E2", basemodel.FieldOrgId: "org_mine"}},
		{Number: 3, Fields: dmodel.DynamicFields{"name": "c", FieldExternalId: "E1", basemodel.FieldOrgId: "org_mine"}},
		{Number: 4, Fields: dmodel.DynamicFields{"name": "d"}},
	}

	result, err := RunBulkCreate(ownerContext(), domSvc, rows, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Data.UpdatedCount)
	assert.Equal(t, 2, result.Data.CreatedCount)
	assert.Equal(t, 4, result.Data.TotalRows)
	require.Len(t, result.Data.Errors, 1)
	assert.Equal(t, 3, result.Data.Errors[0].Row)
	assert.Equal(t, ErrRowDuplicateExternalId, result.Data.Errors[0].Code)
	assert.Equal(t, 1, result.Data.Errors[0].Params["first_row"])

	require.Len(t, domSvc.history, 3)
	assert.Equal(t, "stored_1", domSvc.history[0]["id"], "the stored row is updated by its id")
	assert.Equal(t, "v7", domSvc.history[0][basemodel.FieldEtag], "a versioned row is updated with its stored etag")
	assert.Nil(t, domSvc.history[1]["id"])
	assert.NotNil(t, repo.searchGraph, "existing keys are looked up before the transaction")
}

func TestRunBulkCreateRefusesToUpdateAnotherOrgsRecord(t *testing.T) {
	domSvc := newFakeDomainService(newDedupSchema())
	domSvc.repo.(*stubRepository).searchItems = []dmodel.DynamicFields{
		{"id": "stored_1", FieldSourceSystem: "manual", FieldExternalId: "E1", basemodel.FieldOrgId: "org_other"},
	}
	rows := []BulkRow{
		{Number: 1, Fields: dmodel.DynamicFields{"name": "a", FieldExternalId: "E1", basemodel.FieldOrgId: "org_mine"}},
	}

	result, err := RunBulkCreate(ownerContext(), domSvc, rows, nil)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Data.AffectedCount)
	require.Len(t, result.Data.Errors, 1)
	assert.Equal(t, ErrRowExternalIdOtherOrg, result.Data.Errors[0].Code)
	assert.Empty(t, domSvc.history)
}

func TestRunBulkCreateMatchesTheSchemaDefaultSourceSystem(t *testing.T) {
	writer := &bulkWriter{schema: newDedupSchema()}

	key, ok := writer.dedupKeyOf(dmodel.DynamicFields{FieldExternalId: "E1"})

	require.True(t, ok)
	assert.Equal(t, dedupKey("manual", "E1"), key)
}

func TestCreateBulkChecksPermissionOnce(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema())

	result, err := appSvc.CreateBulk(deniedContext(), BulkCreateCommand{"items": []any{map[string]any{"name": "a"}}})

	require.NoError(t, err)
	assert.Greater(t, result.ClientErrors.Count(), 0)
	assert.Equal(t, 0, domSvc.calls)
}

func TestCreateBulkStampsTheResolvedOrgOnEveryRow(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.CreateBulk(memberContext(), BulkCreateCommand{
		basemodel.FieldOrgId: "org_mine",
		"items":              []any{map[string]any{"name": "a", basemodel.FieldOrgId: "org_other"}},
	})

	require.NoError(t, err)
	require.Equal(t, 0, result.ClientErrors.Count(), result.ClientErrors)
	assert.Equal(t, 1, result.Data.CreatedCount)
	assert.Equal(t, "org_mine", domSvc.seen[basemodel.FieldOrgId])
}

func TestCreateBulkHonoursTheAllowList(t *testing.T) {
	_, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{CrudActions: []CrudAction{CrudActionCreate}})

	result, err := appSvc.CreateBulk(ownerContext(), BulkCreateCommand{"items": []any{}})

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
}
