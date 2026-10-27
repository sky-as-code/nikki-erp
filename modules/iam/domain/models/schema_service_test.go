package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// iam_service is a view over iam_users: every field it declares must be the user column of the
// same name and type, or the shared repository would read and write a different column.
func TestServiceSchemaIsASubsetOfUser(t *testing.T) {
	requireBaseSchemasRegistered(t)
	user := UserSchemaBuilder().Build()
	service := ServiceSchemaBuilder().Build()

	assert.True(t, service.SharesTable())
	assert.Equal(t, user.TableName(), service.TableName())
	assert.Equal(t, user.PrimaryKeys(), service.PrimaryKeys())
	for _, name := range service.FieldNames() {
		serviceField, _ := service.Field(name)
		if serviceField.IsVirtual() {
			continue
		}
		userField, ok := user.Field(name)
		require.True(t, ok, "field %q is not a user column", name)
		assert.Equal(t, userField.DataType().String(), serviceField.DataType().String(), "field %q", name)
	}
}

func TestServiceSchemaOmitsHumanOnlyFields(t *testing.T) {
	requireBaseSchemasRegistered(t)
	service := ServiceSchemaBuilder().Build()
	for _, name := range []string{
		UserFieldEmail, UserFieldAvatarFileId, UserFieldType, UserFieldOrgUnitId, UserFieldIsOwner,
	} {
		_, exists := service.Field(name)
		assert.False(t, exists, "iam_service must not expose %q", name)
	}
}
