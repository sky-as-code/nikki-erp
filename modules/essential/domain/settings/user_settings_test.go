package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// A settings schema describes values; it must not own a table. PrimaryKeys is populated only by
// populateDbMetadata, which runs only under ShouldBuildDb, so an empty one is the check.
func TestUserSettingsSchema_IsMetadataOnly(t *testing.T) {
	schema := UserSettingsSchemaBuilder().Build()

	assert.Empty(t, schema.PrimaryKeys(),
		"a settings schema must not declare should_build_db: settings owns the only tables")
	assert.Empty(t, schema.TableName())
}

// allow_override is what the settings module reads to decide whether a tenant may lock a setting.
// Absent metadata reads as overridable, so a dropped Metadata() call would silently change the
// product's behaviour rather than failing.
func TestUserSettingsSchema_DeclaresAllowOverride(t *testing.T) {
	schema := UserSettingsSchemaBuilder().Build()

	for _, name := range []string{UserSettingThemeMode, UserSettingLanguage} {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		allow, found := field.MetadataValue("allow_override")
		assert.True(t, found, "%s must declare allow_override explicitly", name)
		assert.Equal(t, true, allow, name)
	}
}

// BR §52 requires three theme modes. The frontend's type could not represent "auto" before this
// work, so the enum is pinned here rather than trusted.
func TestUserSettingsSchema_ThemeModeAcceptsAuto(t *testing.T) {
	schema := UserSettingsSchemaBuilder().Build()

	field, ok := schema.Field(UserSettingThemeMode)
	require.True(t, ok)

	values, ok := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues].([]string)
	require.True(t, ok, "theme_mode must be an enum_string")
	assert.ElementsMatch(t, []string{ThemeModeLight, ThemeModeDark, ThemeModeAuto}, values)

	defaultValue := field.Default()
	require.NotNil(t, defaultValue, "theme_mode must default")
	require.NotNil(t, defaultValue.Get())
	assert.Equal(t, ThemeModeAuto, *defaultValue.Get(),
		"a user who expressed no preference has not asked for either fixed theme")
}

// The language enum must be exactly the locale set the application ships, because there is no
// fallback language: an unsupported locale renders every key as its raw namespace:key.
func TestUserSettingsSchema_LanguageMatchesSupportedLocales(t *testing.T) {
	schema := UserSettingsSchemaBuilder().Build()

	field, ok := schema.Field(UserSettingLanguage)
	require.True(t, ok)

	values, ok := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues].([]string)
	require.True(t, ok, "language must be an enum_string")
	assert.ElementsMatch(t, SupportedLanguages, values)

	// BR §53: vi-VN must not be hardcoded as the business default. The user context falls back to a
	// display locale, but the schema itself declares no default language.
	assert.Nil(t, field.Default(),
		"language must not carry a default; BR §53 forbids a hardcoded business default")
}

// Every setting must carry a description, and it must be a $ref rather than literal text: the
// sentence explaining a setting is shown in the settings pane and has to be translated like
// everything else. The key convention is settings_desc.<setting name>, which is what makes the
// language files greppable from the setting name alone.
func TestSettingsSchemas_DescribeEverySettingByRef(t *testing.T) {
	schemas := map[string]*dmodel.ModelSchema{
		UserSettingsSchemaName: UserSettingsSchemaBuilder().Build(),
		OrgSettingsSchemaName:  OrgSettingsSchemaBuilder().Build(),
	}

	for schemaName, schema := range schemas {
		for _, name := range schema.FieldNames() {
			field, ok := schema.Field(name)
			require.True(t, ok, name)

			description := field.Description()
			require.NotEmpty(t, description,
				"%s.%s must carry a description", schemaName, name)

			ref, isRef := description[model.LanguageCodeRef]
			assert.True(t, isRef,
				"%s.%s must describe itself by $ref, not literal text", schemaName, name)
			assert.Equal(t, "settings_desc."+name, ref,
				"the description key convention is settings_desc.<setting name>")
		}
	}
}

// The organization level describes the organization's own conventions. Declaring one of these at
// the user level instead would let each person quietly redefine what a shared report means.
func TestOrgSettingsSchema_DeclaresTheOrganizationConventions(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.ElementsMatch(t,
		[]string{OrgSettingSystemLocale, OrgSettingSystemTimezone, OrgSettingDefaultCurrency},
		schema.FieldNames())
	assert.Empty(t, schema.PrimaryKeys(),
		"a settings schema must not declare should_build_db: settings owns the only tables")
}

// A user's timezone is deliberately separate from the organization's: the organization's zone is
// what a shared report is stamped with, the user's is how that report is shown to whoever reads it.
// Collapsing the two would make one of those two things wrong for anyone working remotely.
func TestUserAndOrgTimezonesAreDistinctSettings(t *testing.T) {
	userSchema := UserSettingsSchemaBuilder().Build()
	orgSchema := OrgSettingsSchemaBuilder().Build()

	_, userHasTimezone := userSchema.Field(UserSettingTimezone)
	assert.True(t, userHasTimezone)

	// The names must differ. The settings_records unique key carries no level, so one module
	// declaring the same name at two levels collides on write for an owner holding both.
	assert.NotEqual(t, UserSettingTimezone, OrgSettingSystemTimezone)
	for _, name := range orgSchema.FieldNames() {
		_, clashes := userSchema.Field(name)
		assert.False(t, clashes,
			"setting '%s' is declared at both the user and org level of this module", name)
	}
}
