package timezone

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Every id must be loadable, or a stored setting would name a zone the server cannot convert with.
func TestEveryIdLoads(t *testing.T) {
	for _, id := range Ids {
		_, err := time.LoadLocation(id)
		assert.NoError(t, err, id)
	}
	assert.Len(t, known, len(Ids), "ids must be unique")
}

func TestSearchMatchesIdAndLabel(t *testing.T) {
	labels := map[string]string{"Asia/Ho_Chi_Minh": "Châu Á/Hồ_Chí_Minh"}

	items, total := Search("ho chi", labels, 0, 10)
	assert.Equal(t, 1, total)
	assert.Equal(t, Zone{Code: "Asia/Ho_Chi_Minh", Label: "Châu Á/Hồ_Chí_Minh"}, items[0])

	_, total = Search("HỒ_CHÍ", labels, 0, 10)
	assert.Equal(t, 1, total)

	items, total = Search("", nil, 1, 10)
	assert.Equal(t, len(Ids), total)
	assert.Len(t, items, 10)
	assert.Equal(t, Ids[10], items[0].Code)

	items, _ = Search("", nil, 1000, 10)
	assert.Empty(t, items)
}
