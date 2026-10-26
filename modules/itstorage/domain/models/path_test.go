package models

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

func TestSchemas_Build(t *testing.T) {
	space := SpaceSchemaBuilder().Build()
	assert.Equal(t, "itstorage_spaces", space.TableName())
	assert.NotEmpty(t, space.PrimaryKeys())

	file := FileSchemaBuilder().Build()
	assert.Equal(t, "itstorage_files", file.TableName())
	required := []string{FileFieldName, FileFieldMimeType, FileFieldSizeBytes, FileFieldSpaceId, FileFieldObjectKey}
	for _, name := range required {
		_, ok := file.Field(name)
		assert.True(t, ok, "itstorage_file must declare %q", name)
	}
}

func TestSpaceName_RejectsUnsafeKeyCharacters(t *testing.T) {
	field, _ := SpaceSchemaBuilder().Build().Field(SpaceFieldName)
	_, cErr := field.Validate("01JWNY20G23KD4RV5VWYABQYHD")
	assert.Nil(t, cErr)
	_, cErr = field.Validate("a/b")
	assert.NotNil(t, cErr)
	_, cErr = field.Validate("has space")
	assert.NotNil(t, cErr)
}

func TestFileName_AcceptsUnicodeButNotSeparators(t *testing.T) {
	field, _ := FileSchemaBuilder().Build().Field(FileFieldName)
	_, cErr := field.Validate("Ảnh đại diện.png")
	assert.Nil(t, cErr)
	_, cErr = field.Validate("../etc/passwd")
	assert.NotNil(t, cErr)
}

func TestParseFilePath(t *testing.T) {
	parsed, ok := ParseFilePath("01KNV107H8P0V2CB1R9T9ZNMFE/01JWNY20G23KD4RV5VWYABQYHD/01K0000000000000000000000A.png")
	assert.True(t, ok)
	assert.Equal(t, "01KNV107H8P0V2CB1R9T9ZNMFE/01JWNY20G23KD4RV5VWYABQYHD", parsed.SpacePath)
	assert.Equal(t, "01K0000000000000000000000A", parsed.FileId)
	assert.Equal(t, "01K0000000000000000000000A.png", parsed.Leaf)

	parsed, ok = ParseFilePath("root/01K0000000000000000000000A")
	assert.True(t, ok)
	assert.Equal(t, "01K0000000000000000000000A", parsed.Leaf)

	for _, bad := range []string{
		"", "01K0000000000000000000000A.png", "root/", "root/not-an-id.png", "/01K0000000000000000000000A",
	} {
		_, ok := ParseFilePath(bad)
		assert.False(t, ok, bad)
	}
}

func TestPathHelpers(t *testing.T) {
	assert.Equal(t, "root", JoinPath("", "root"))
	assert.Equal(t, "root/org", JoinPath("root", "org"))
	assert.True(t, IsUnderPath("root/org", "root"))
	assert.False(t, IsUnderPath("rootx/org", "root"))
	assert.Equal(t, ".png", ExtensionOf("Photo.PNG"))
	assert.Equal(t, "", ExtensionOf("noext"))
	assert.Equal(t, "", ExtensionOf("weird.p n g"))
	assert.Equal(t, "_/S/F.mp4", BuildObjectKey("", "S", "F", ".mp4"))
	assert.Equal(t, "T/S/F", BuildObjectKey("T", "S", "F", ""))
}

func TestMimeFamilyMatches(t *testing.T) {
	assert.True(t, MimeFamilyMatches("image/jpeg", "image/png"))
	assert.True(t, MimeFamilyMatches("video/mp4", "video/mp4"))
	assert.True(t, MimeFamilyMatches("application/json", "text/plain; charset=utf-8"))
	assert.True(t, MimeFamilyMatches("text/csv", "text/plain"))
	assert.True(t, MimeFamilyMatches("application/x-custom", "application/octet-stream"))
	assert.False(t, MimeFamilyMatches("image/png", "application/pdf"))
	assert.False(t, MimeFamilyMatches("image/png", "text/plain"))
	assert.False(t, MimeFamilyMatches("video/mp4", "image/jpeg"))
}
