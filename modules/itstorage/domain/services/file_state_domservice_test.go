package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
)

func fileWith(fields dmodel.DynamicFields) *models.File {
	return models.NewFileFrom(fields)
}

func TestUnusableReason(t *testing.T) {
	org := model.Id("ORG1")
	other := model.Id("ORG2")
	cases := map[string]struct {
		file  *models.File
		orgId *model.Id
		ok    bool
	}{
		"missing":          {nil, &org, false},
		"other org":        {fileWith(dmodel.DynamicFields{"org_id": other, "status": "pending"}), &org, false},
		"org file, no org": {fileWith(dmodel.DynamicFields{"org_id": org, "status": "pending"}), nil, false},
		"in use":           {fileWith(dmodel.DynamicFields{"org_id": org, "status": "in_use"}), &org, false},
		"released": {fileWith(dmodel.DynamicFields{
			"org_id": org, "status": "pending", "in_use_at": "2026-01-01T00:00:00Z",
		}), &org, false},
		"pending org file":    {fileWith(dmodel.DynamicFields{"org_id": org, "status": "pending"}), &org, true},
		"pending tenant file": {fileWith(dmodel.DynamicFields{"status": "pending"}), nil, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ok, unusableReason(tc.file, tc.orgId) == "")
		})
	}
}

func TestUniqueIds(t *testing.T) {
	assert.Equal(t, []model.Id{"A", "B"}, uniqueIds([]model.Id{"B", "", "A", "B"}))
}
