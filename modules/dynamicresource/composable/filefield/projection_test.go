package filefield

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

var testFields = []FileField{{Name: "avatar", KeyField: "avatar_file_key", UrlField: "avatar_file_url"}}

func TestAppendHiddenKeys_DefaultProjectionGainsTheKey(t *testing.T) {
	query := dmodel.DynamicFields{}
	defaults := []string{"avatar_url", "email"}

	added := appendHiddenKeys(query, testFields, defaults)

	assert.Equal(t, []string{"avatar_file_key"}, added)
	assert.Equal(t, []string{"avatar_url", "email", "avatar_file_key"}, query[paramFields])
	assert.Equal(t, []string{"avatar_url", "email"}, defaults, "the shared default list must not be written into")
}

func TestAppendHiddenKeys_LeavesFullReadsAlone(t *testing.T) {
	assert.Nil(t, appendHiddenKeys(dmodel.DynamicFields{}, testFields, nil), "no projection is the whole record")
	assert.Nil(t, appendHiddenKeys(dmodel.DynamicFields{paramFields: []any{"*", "orgs.name"}}, testFields, nil))
	assert.Nil(t, appendHiddenKeys(dmodel.DynamicFields{paramFields: []string{"avatar_file_key"}}, testFields, nil))
}

func TestStripHiddenKeys_RemovesOnlyWhatWasAdded(t *testing.T) {
	rows := []dmodel.DynamicFields{{"email": "a@b.c", "avatar_file_key": "k", "avatar_file_url": "u"}}

	desired := stripHiddenKeys(rows, []string{"email", "avatar_file_key"}, []string{"avatar_file_key"})

	assert.Equal(t, []string{"email"}, desired)
	assert.Equal(t, dmodel.DynamicFields{"email": "a@b.c", "avatar_file_url": "u"}, rows[0])
}
