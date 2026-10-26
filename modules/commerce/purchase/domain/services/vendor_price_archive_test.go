package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Which set_archived requests trigger revalidation. The two directions are not symmetric: archiving
// is always allowed, while unarchiving must re-check the vendor, product and unit that may have
// become unusable while the row was retired. A request misread as archiving skips the check and
// restores a price that could not be created.

func archivedParams(value any) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		basemodel.FieldId:         "01VPP",
		basemodel.FieldIsArchived: value,
	}
}

func TestUnarchivingIsRecognised(t *testing.T) {
	assert.True(t, isUnarchiving(archivedParams(false)),
		"is_archived=false is a request to bring the row back, and must revalidate")
}

func TestArchivingIsNotRevalidated(t *testing.T) {
	assert.False(t, isUnarchiving(archivedParams(true)),
		"withdrawing an offer breaks nothing downstream and needs no check")
}

// A pointer is accepted because the params may arrive either shape, and reading a *bool as
// unrecognised would skip the check on a genuine unarchive.
func TestAPointerFlagIsReadTheSameWay(t *testing.T) {
	no, yes := false, true

	assert.True(t, isUnarchiving(archivedParams(&no)))
	assert.False(t, isUnarchiving(archivedParams(&yes)))
}

// An absent flag is not an unarchive: the engine builds its command with a nil flag and the
// repository leaves the stored value alone, so nothing is being restored.
func TestAnAbsentFlagIsNotAnUnarchive(t *testing.T) {
	assert.False(t, isUnarchiving(dmodel.DynamicFields{basemodel.FieldId: "01VPP"}))
	assert.False(t, isUnarchiving(archivedParams(nil)))

	var missing *bool
	assert.False(t, isUnarchiving(archivedParams(missing)),
		"a nil pointer carries no instruction either")
}

// An unreadable value is treated as "not an unarchive", so the base call reports the bad parameter
// rather than this code inventing a direction.
func TestAnUnreadableFlagIsNotTreatedAsAnUnarchive(t *testing.T) {
	assert.False(t, isUnarchiving(archivedParams("false")))
}
