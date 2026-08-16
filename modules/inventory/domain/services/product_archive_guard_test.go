package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Whether a product may be archived.
//
// The line these draw is between stock that is still live and stock that is merely remembered. A
// variant with goods on the shelf, a promise against them, or work in flight cannot go; one whose
// only trace is completed movement can.

// stubUsageReader answers with fixed usage, so a test states the stock situation directly rather
// than assembling quants and moves to imply it.
type stubUsageReader struct {
	usages map[string]itStock.ProductUsage

	// batchCalls counts how many times the batch read was made, which is how the "checked as one
	// set" property is asserted.
	batchCalls int
}

func (this *stubUsageReader) GetProductUsage(
	_ corectx.Context, query itStock.GetProductUsageQuery,
) (*itStock.GetProductUsageResult, error) {
	return &itStock.GetProductUsageResult{
		HasData: true,
		Data:    itStock.GetProductUsageResultData{Usage: this.usages[query.VariantId]},
	}, nil
}

func (this *stubUsageReader) GetProductUsageBatch(
	_ corectx.Context, query itStock.GetProductUsageBatchQuery,
) (*itStock.GetProductUsageBatchResult, error) {
	this.batchCalls++

	usages := map[string]itStock.ProductUsage{}
	for _, id := range query.VariantIds {
		usages[id] = this.usages[id]
	}
	return &itStock.GetProductUsageBatchResult{
		HasData: true,
		Data:    itStock.GetProductUsageBatchResultData{Usages: usages},
	}, nil
}

func usageWithOnHand(quantity int64) itStock.ProductUsage {
	return itStock.ProductUsage{OnHandQuantity: decimal.NewFromInt(quantity)}
}

// TS-PROD-10: a variant still holding stock cannot be archived, and archiving generates no
// movement to make it go away.
func TestVariantWithStockCannotBeArchived(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertVariantArchivable(usageWithOnHand(100), vErrs)

	assert.Equal(t, 1, vErrs.Count(), "on-hand stock blocks the archive")
}

// TS-PROD-11: history alone does not block. A variant that has been moved but holds nothing now
// archives fine, and the completed records keep resolving it afterwards.
func TestVariantWithOnlyHistoryCanBeArchived(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertVariantArchivable(itStock.ProductUsage{}, vErrs)

	assert.Equal(t, 0, vErrs.Count(),
		"completed movement is history, and history never blocks archiving")
}

// Each blocker is reported, not just the first, so a user clearing them sees the whole list at
// once rather than discovering the next one after fixing each.
func TestVariantReportsEveryBlockingReason(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertVariantArchivable(itStock.ProductUsage{
		OnHandQuantity:    decimal.NewFromInt(5),
		ReservedQuantity:  decimal.NewFromInt(2),
		OpenMoveCount:     1,
		OpenTransferCount: 1,
	}, vErrs)

	assert.Equal(t, 4, vErrs.Count(), "all four reasons are reported together")
}

// Reservation blocks on its own. It is stock promised to someone else, and archiving the product
// would strand that promise even though nothing needs physically moving.
func TestVariantWithOnlyReservationCannotBeArchived(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertVariantArchivable(
		itStock.ProductUsage{ReservedQuantity: decimal.NewFromInt(3)}, vErrs)

	assert.Equal(t, 1, vErrs.Count())
}

// Unarchiving is never guarded: it puts a product back into use and strands nothing. Guarding it
// would make a variant archived by mistake impossible to recover.
func TestUnarchivingIsNotGuarded(t *testing.T) {
	reader := &stubUsageReader{
		usages: map[string]itStock.ProductUsage{testVariantAId: usageWithOnHand(100)},
	}
	vErrs := &ft.ClientErrors{}

	err := GuardVariantArchive(callerContext(), reader, testVariantAId, false, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 0, vErrs.Count(), "restoring a product to use strands nothing")
}

// TS-PROD-12: one variant with stock blocks the whole template. The line is archived as a unit or
// not at all.
func TestTemplateArchiveBlockedByASingleVariant(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertTemplateArchivable(
		map[string]itStock.ProductUsage{
			testVariantAId: {},
			testVariantBId: usageWithOnHand(5),
		},
		map[string]string{testVariantBId: "SKU-B"},
		vErrs,
	)

	assert.Equal(t, 1, vErrs.Count(), "the clean variant passes, the stocked one blocks")
	assert.Contains(t, vErrs.ToError().Error(), "SKU-B",
		"the offending variant is named, so a user is not left hunting for it")
}

// A template whose variants are all clear archives.
func TestTemplateArchivePassesWhenEveryVariantIsClear(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	AssertTemplateArchivable(
		map[string]itStock.ProductUsage{testVariantAId: {}, testVariantBId: {}},
		map[string]string{},
		vErrs,
	)

	assert.Equal(t, 0, vErrs.Count())
}

// The whole set is read in one call, before anything is written.
//
// This is the property that keeps a rejected archive from leaving half a product line archived:
// the cascade writes as it walks, so a per-variant check made during that walk would already have
// archived the variants it passed.
func TestTemplateGuardReadsEveryVariantInOneCall(t *testing.T) {
	repo := &stubVariantRepository{variants: []dmodel.DynamicFields{
		archivedVariant(testVariantAId, false, nil),
		archivedVariant(testVariantBId, false, nil),
	}}
	useVariantEngine(t, repo)

	reader := &stubUsageReader{
		usages: map[string]itStock.ProductUsage{testVariantBId: usageWithOnHand(5)},
	}
	vErrs := &ft.ClientErrors{}

	err := GuardTemplateArchive(callerContext(), reader, testTemplateId, true, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 1, reader.batchCalls, "one batch read covers the whole line, not one per variant")
	assert.Equal(t, 1, vErrs.Count())
}

// An already-archived variant is not being withdrawn by this operation, so whatever it holds is a
// pre-existing condition rather than something this archive would strand. Counting it would make
// the template permanently unarchivable.
func TestTemplateGuardSkipsAlreadyArchivedVariants(t *testing.T) {
	repo := &stubVariantRepository{variants: []dmodel.DynamicFields{
		archivedVariant(testVariantAId, true, nil),
	}}
	useVariantEngine(t, repo)

	reader := &stubUsageReader{
		usages: map[string]itStock.ProductUsage{testVariantAId: usageWithOnHand(99)},
	}
	vErrs := &ft.ClientErrors{}

	err := GuardTemplateArchive(callerContext(), reader, testTemplateId, true, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 0, vErrs.Count(),
		"a variant already out of the working set is not being withdrawn again")
}

// useVariantEngine points the engine lookup at a stub variant repository.
func useVariantEngine(t *testing.T, repo *stubVariantRepository) {
	t.Helper()

	original := engineFor
	t.Cleanup(func() { engineFor = original })
	engineFor = func(_ string) (drif.DynamicResourceEngine, error) {
		return &stubEngine{repo: repo}, nil
	}
}
