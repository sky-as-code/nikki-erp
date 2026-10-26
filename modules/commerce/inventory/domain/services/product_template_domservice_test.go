package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The archive-cascade rules: which variants a template's archive takes down, and which its
// unarchive brings back.

// Unarchiving a template restores the variants it cascaded to and leaves the ones a user archived
// deliberately alone. A plain boolean cannot express that, which is why archive_source exists.
func TestCascadeSkipsUserArchivedVariantOnUnarchive(t *testing.T) {
	testCases := []struct {
		name        string
		source      *models.ArchiveSource
		unarchiving bool
		wantSkipped bool
	}{
		{
			name:        "a cascade-archived variant comes back with its template",
			source:      ptr(models.ArchiveSourceTemplateCascade),
			unarchiving: true,
			wantSkipped: false,
		},
		{
			name:        "a user-archived variant stays archived",
			source:      ptr(models.ArchiveSourceUser),
			unarchiving: true,
			wantSkipped: true,
		},
		{
			name:        "a sync-archived variant stays archived",
			source:      ptr(models.ArchiveSourceSystemSync),
			unarchiving: true,
			wantSkipped: true,
		},
		{
			name:        "archiving ignores the source entirely",
			source:      ptr(models.ArchiveSourceUser),
			unarchiving: false,
			wantSkipped: false,
		},
		{
			// A variant archived before archive_source existed carries no stamp; leaving it archived
			// beats resurrecting a deliberately withdrawn product.
			name:        "an unstamped variant stays archived",
			source:      nil,
			unarchiving: true,
			wantSkipped: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			variant := models.NewProductVariant()
			variant.SetIsArchived(ptr(testCase.unarchiving))
			variant.SetArchiveSource(testCase.source)

			assert.Equal(t, testCase.wantSkipped, ShouldSkipCascade(variant, !testCase.unarchiving))
		})
	}
}

// A variant already in the target state needs no write at all, whatever archived it.
func TestCascadeSkipsVariantAlreadyInTargetState(t *testing.T) {
	archived := models.NewProductVariant()
	archived.SetIsArchived(ptr(true))
	archived.SetArchiveSource(ptr(models.ArchiveSourceTemplateCascade))

	assert.True(t, ShouldSkipCascade(archived, true), "already archived, nothing to do")

	unarchived := models.NewProductVariant()
	unarchived.SetIsArchived(ptr(false))

	assert.True(t, ShouldSkipCascade(unarchived, false), "already unarchived, nothing to do")
}

// The counterpart to the already-in-target-state case: a variant the template's archive took
// down is exactly what unarchiving the template must bring back.
func TestCascadeRestoresCascadeArchivedVariant(t *testing.T) {
	cascaded := models.NewProductVariant()
	cascaded.SetIsArchived(ptr(true))
	cascaded.SetArchiveSource(ptr(models.ArchiveSourceTemplateCascade))

	assert.False(t, ShouldSkipCascade(cascaded, false))
}

// The stamp is what lets a later unarchive tell its own cascade apart from a deliberate archive,
// so the delta must carry it on the way down and clear it on the way back up.
func TestCascadeArchiveFieldsStampsTheSource(t *testing.T) {
	archiving := CascadeArchiveFields("01VARIANT", true)

	assert.Equal(t, "01VARIANT", archiving[models.ProductVariantFieldId])
	assert.Equal(t, true, archiving[basemodel.FieldIsArchived])
	assert.Equal(t, models.ArchiveSourceTemplateCascade.String(),
		archiving[models.ProductVariantFieldArchiveSource])

	unarchiving := CascadeArchiveFields("01VARIANT", false)

	assert.Equal(t, false, unarchiving[basemodel.FieldIsArchived])
	assert.Nil(t, unarchiving[models.ProductVariantFieldArchiveSource],
		"the stamp is cleared, so a later archive can set its own reason")
}

func ptr[T any](v T) *T {
	return &v
}
