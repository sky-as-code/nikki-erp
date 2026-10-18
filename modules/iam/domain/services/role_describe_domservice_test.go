package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// The repository may hand a hydrated LangJson leaf back either typed or as the raw decoded map,
// so both shapes must flatten to the default-language label.
func TestDescribedEntitlementFrom_FlattensLangJsonNames(t *testing.T) {
	fields := dmodel.DynamicFields{
		domain.EntitlementEdgeResource: dmodel.DynamicFields{
			domain.ResourceFieldName: model.LangJson{"en-US": "IAM User", "vi-VN": "Người dùng IAM"},
		},
		domain.EntitlementEdgeAction: dmodel.DynamicFields{
			domain.ActionFieldName: map[string]interface{}{"vi-VN": "Xem", "en-US": "Read"},
		},
	}

	described := describedEntitlementFrom(model.Id("01JWNZ3XVWDP73JGHRRBFAHQYJ"), fields)

	require.NotNil(t, described.ResourceName)
	require.NotNil(t, described.ActionName)
	assert.Equal(t, "IAM User", *described.ResourceName)
	assert.Equal(t, "Read", *described.ActionName)
}

func TestDescribedEntitlementFrom_MissingEdgeLeavesNameNil(t *testing.T) {
	described := describedEntitlementFrom(model.Id("01JWNZ3XVWDP73JGHRRBFAHQYJ"), dmodel.DynamicFields{
		basemodel.FieldId: "01JWNZ3XVWDP73JGHRRBFAHQYJ",
	})

	assert.Nil(t, described.ResourceName)
	assert.Nil(t, described.ActionName)
}
