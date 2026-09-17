package baserepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const versionedRepoSchemaName = "test_baserepo_versioned"

// versionedRepo builds a repository over a schema that extends versioned_model, which is what
// gives it an etag column.
func versionedRepo(t *testing.T) *BaseDynamicRepositoryImpl {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if registry.Get(versionedRepoSchemaName) == nil {
		require.NoError(t, dmodel.RegisterSchemaB(
			dmodel.DefineModel(versionedRepoSchemaName).
				TableName("test_baserepo_versioneds").
				ShouldBuildDb().
				Field(dmodel.DefineField().Name("id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
				Field(dmodel.DefineField().Name("sku").
					DataType(dmodel.FieldDataTypeString(1, 100)).RequiredForCreate()).
				Extend(basemodel.VersionedModelSchemaBuilder())))
	}

	return &BaseDynamicRepositoryImpl{schema: registry.Get(versionedRepoSchemaName)}
}

// The rule this whole change exists for: a row read through an explicit projection must still carry
// the etag that writing it back checks against. No client asks for etag -- the field picker hides
// system fields deliberately -- so if the repository does not add it, nothing does.
func TestEnsureSystemColumns_InjectsEtagWhenSchemaIsVersioned(t *testing.T) {
	repo := versionedRepo(t)

	columns := repo.ensureSystemColumns([]string{"sku"})

	assert.Contains(t, columns, basemodel.FieldEtag)
	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "sku")
}

// A schema that never extended versioned_model has no etag column, and selecting one would be a
// SQL error rather than a missing value.
func TestEnsureSystemColumns_SkipsEtagWhenSchemaIsNotVersioned(t *testing.T) {
	repo := virtualRepo(t)

	columns := repo.ensureSystemColumns([]string{"sku"})

	assert.NotContains(t, columns, basemodel.FieldEtag)
	assert.Contains(t, columns, "id")
}

// An empty list is SELECT *, which already carries every column. Turning it into an explicit list
// here would narrow the query instead of widening it -- the one way this function can lose data.
func TestEnsureSystemColumns_EmptyInputStaysEmpty(t *testing.T) {
	repo := versionedRepo(t)

	assert.Empty(t, repo.ensureSystemColumns(nil))
	assert.Empty(t, repo.ensureSystemColumns([]string{}))
}

func TestEnsureSystemColumns_DoesNotDuplicateAnExplicitEtag(t *testing.T) {
	repo := versionedRepo(t)

	columns := repo.ensureSystemColumns([]string{"sku", basemodel.FieldEtag, "id"})

	assert.ElementsMatch(t, []string{"id", "sku", basemodel.FieldEtag}, columns)
}
