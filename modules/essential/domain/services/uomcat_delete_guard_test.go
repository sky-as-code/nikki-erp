package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// A UoM Category may only be hard-deleted once it holds no UoM: the UoM foreign key already
// refuses it, and the guard reports that as a business violation instead of a 500.

// stubUomSearcher answers every search with the same rows, which is all the guard reads: it asks
// whether any UoM exists in the category, not which.
type stubUomSearcher struct {
	rows []dmodel.DynamicFields
}

func (this *stubUomSearcher) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.rows},
		HasData: len(this.rows) > 0,
	}, nil
}

var _ models.UomSearcher = (*stubUomSearcher)(nil)

const testUomCatId = "uomcat-1"

func guardContext() corectx.Context {
	return corectx.NewRequestContextM(context.Background(), "essential")
}

func TestUomCatWithUomsCannotBeDeleted(t *testing.T) {
	repo := &stubUomSearcher{rows: []dmodel.DynamicFields{{models.UomFieldId: "uom-1"}}}
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertUomCatDeletable(guardContext(), repo, testUomCatId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "a UoM still belongs to the category")
}

func TestEmptyUomCatCanBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertUomCatDeletable(guardContext(), &stubUomSearcher{}, testUomCatId, vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

// An empty id is a malformed request, not a business refusal: schema validation reports it.
func TestUomCatGuardIgnoresEmptyId(t *testing.T) {
	repo := &stubUomSearcher{rows: []dmodel.DynamicFields{{models.UomFieldId: "uom-1"}}}
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertUomCatDeletable(guardContext(), repo, "", vErrs))

	assert.Equal(t, 0, vErrs.Count())
}
