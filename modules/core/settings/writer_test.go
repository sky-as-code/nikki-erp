package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type recordingService struct {
	calls []SetSettingsCommand
}

func (this *recordingService) SetSettings(
	ctx corectx.Context, level string, cmd SetSettingsCommand,
) (*SetSettingsResult, error) {
	this.calls = append(this.calls, cmd)
	return &SetSettingsResult{Data: SetSettingsResultData{Updated: len(cmd.Items)}, HasData: true}, nil
}

func testSchema() *dmodel.ModelSchema {
	return dmodel.ParseModelJson(`{
		"name": "test_user_settings",
		"fields": [{ "name": "theme", "data_type": { "type": "enum_string", "values": ["light", "dark"] } }]
	}`).Build()
}

func newTestWriter(rules ...SettingsRule) (*ModuleSettingsWriter, *recordingService) {
	service := &recordingService{}
	schemas := map[string]*dmodel.ModelSchema{LevelUser: testSchema()}
	return NewModuleSettingsWriter("test", schemas, service, rules...), service
}

func update(t *testing.T, writer *ModuleSettingsWriter, level string, value any) *SetSettingsResult {
	ctx := corectx.NewRequestContext(context.Background())
	result, err := writer.Update(ctx, UpdateSettingsCommand{
		Level: level, Items: []SetSettingItem{{Name: "theme", Value: value}},
	})
	require.NoError(t, err)
	return result
}

func TestModuleSettingsWriter_PersistsUnderItsOwnModuleKey(t *testing.T) {
	writer, service := newTestWriter()
	result := update(t, writer, LevelUser, "dark")
	assert.True(t, result.HasData)
	require.Len(t, service.calls, 1)
	assert.Equal(t, "test", service.calls[0].ModuleKey)
}

func TestModuleSettingsWriter_RefusesALevelWithNoSchema(t *testing.T) {
	service := &recordingService{}
	writer := NewModuleSettingsWriter("test", map[string]*dmodel.ModelSchema{}, service)
	result := update(t, writer, LevelUser, "dark")
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, service.calls)
}

func TestModuleSettingsWriter_RefusesAValueOutsideTheSchema(t *testing.T) {
	writer, service := newTestWriter()
	result := update(t, writer, LevelUser, "sepia")
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, service.calls)
}

func TestModuleSettingsWriter_ModuleRuleCanRefuse(t *testing.T) {
	refuse := func(ctx corectx.Context, level string, items []SetSettingItem) (ft.ClientErrors, error) {
		return ft.ClientErrors{*ft.NewValidationError("theme", "test.refused", "refused")}, nil
	}
	writer, service := newTestWriter(refuse)
	result := update(t, writer, LevelUser, "dark")
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, service.calls)
}
