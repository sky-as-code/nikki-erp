package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type stubOrgSettingsReader struct {
	values    map[string]any
	err       error
	moduleKey string
}

func (this *stubOrgSettingsReader) GetOrgSettingValues(
	_ corectx.Context, _ model.Id, moduleKey string,
) (map[string]any, error) {
	this.moduleKey = moduleKey
	return this.values, this.err
}

func TestOrgLanguage(t *testing.T) {
	cases := []struct {
		name   string
		values map[string]any
		want   model.LanguageCode
	}{
		{"unset", map[string]any{}, model.DefaultLanguageCode},
		{"nil map", nil, model.DefaultLanguageCode},
		{"stored", map[string]any{"system_language": "en-US"}, model.LanguageCodeEnUs},
		{"canonicalised", map[string]any{"system_language": "vi_vn"}, model.LanguageCodeViVn},
		{"invalid", map[string]any{"system_language": "not a language"}, model.DefaultLanguageCode},
		{"no region", map[string]any{"system_language": "en"}, model.DefaultLanguageCode},
		{"wrong type", map[string]any{"system_language": 3}, model.DefaultLanguageCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := &stubOrgSettingsReader{values: tc.values}
			got, err := OrgLanguage(nil, reader, "org")
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, "essential", reader.moduleKey)
		})
	}
}

func TestOrgLanguageReadError(t *testing.T) {
	_, err := OrgLanguage(nil, &stubOrgSettingsReader{err: errors.New("down")}, "org")
	assert.Error(t, err)
}
