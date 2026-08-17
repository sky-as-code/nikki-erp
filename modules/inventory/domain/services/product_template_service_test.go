package services

import (
	"context"
	"testing"

	"go.bryk.io/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The archive cascade, from the service side. Which variants it touches is decided by
// ShouldSkipCascade and tested there; what follows covers the wiring the override adds:
// the transaction, the scoped context, and the fact that the writes address variant rows.

const (
	testTemplateId = "01TEMPLATE0000000000000000"
	testCallerId   = "01CALLER00000000000000000A"
)

// stubVariantRepository stands in for the variant engine's repository. Only the three methods the
// cascade uses are implemented; the embedded interface covers the rest so this stays a partial
// double rather than a full reimplementation.
type stubVariantRepository struct {
	drif.DynamicResourceRepository

	variants []dmodel.DynamicFields
	updates  []dmodel.DynamicFields
	// updateContexts records the context each write ran under, so a test can assert the caller's
	// identity and the transaction reached the repository.
	updateContexts []corectx.Context
	updateErr      error
	tranx          *stubTransaction
}

func (this *stubVariantRepository) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.variants},
		HasData: len(this.variants) > 0,
	}, nil
}

func (this *stubVariantRepository) Update(
	ctx corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if this.updateErr != nil {
		return nil, this.updateErr
	}
	this.updates = append(this.updates, data)
	this.updateContexts = append(this.updateContexts, ctx)
	return &dyn.OpResult[dyn.MutateResultData]{HasData: true}, nil
}

func (this *stubVariantRepository) BeginTransaction(_ corectx.Context) (database.DbTransaction, error) {
	this.tranx = &stubTransaction{}
	return this.tranx, nil
}

type stubTransaction struct {
	committed  bool
	rolledBack bool
}

func (this *stubTransaction) Commit() error {
	this.committed = true
	return nil
}

func (this *stubTransaction) Rollback() error {
	this.rolledBack = true
	return nil
}

// stubBaseService stands in for the engine's default resource service, which the override
// delegates the template write to.
type stubBaseService struct {
	drif.DynamicResourceService

	calls        int
	seenContexts []corectx.Context
	clientErrors ft.ClientErrors
	err          error
}

func (this *stubBaseService) SetArchived(
	ctx corectx.Context, _ dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls++
	this.seenContexts = append(this.seenContexts, ctx)
	if this.err != nil {
		return nil, this.err
	}
	return &dyn.OpResult[dyn.MutateResultData]{
		ClientErrors: this.clientErrors,
		HasData:      this.clientErrors.Count() == 0,
	}, nil
}

// callerContext carries an identity, which is what the audit columns are stamped from.
func callerContext() corectx.Context {
	ctx := corectx.NewRequestContextM(context.Background(), "inventory")
	ctx.SetPermissions(corectx.ContextPermissions{UserId: model.Id(testCallerId)})
	return ctx
}

func archivedVariant(id string, isArchived bool, source *models.ArchiveSource) dmodel.DynamicFields {
	fields := dmodel.DynamicFields{
		models.ProductVariantFieldId: id,
		basemodel.FieldIsArchived:    isArchived,
	}
	if source != nil {
		fields[models.ProductVariantFieldArchiveSource] = source.String()
	}
	return fields
}

func archiveParams(isArchived bool) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.ProductTemplateFieldId: testTemplateId,
		basemodel.FieldIsArchived:     isArchived,
	}
}

// stubEngine hands out the stub repository. Only ResourceRepository is reached by the cascade.
type stubEngine struct {
	drif.DynamicResourceEngine

	repo drif.DynamicResourceRepository
}

func (this *stubEngine) ResourceRepository() drif.DynamicResourceRepository {
	return this.repo
}

// runCascadeRaw archives a template against the stub engines, restoring the engine lookup
// afterwards so the swap cannot leak into another test.
func runCascadeRaw(
	t *testing.T, repo *stubVariantRepository, base *stubBaseService, archive bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	t.Helper()

	original := engineFor
	t.Cleanup(func() { engineFor = original })
	engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
		// Two schemas are legitimately reached. The cascade writes variant rows, so its writes
		// must go through the variant engine. The stock guard that runs first asks the quant
		// engine what the variants hold — and is answered here with an error, which the guard
		// treats as "stock is not wired in this deployment" and lets the archive proceed. That
		// keeps these tests about the cascade; the guard has its own.
		switch schemaName {
		case models.ProductVariantSchemaName:
			return &stubEngine{repo: repo}, nil
		case models.StockQuantSchemaName:
			return nil, errors.New("no stock engine in this test")
		}
		require.Failf(t, "unexpected engine", "the cascade reached the '%s' engine", schemaName)
		return nil, nil
	}

	service := &ProductTemplateDomainServiceImpl{DynamicResourceService: base}
	return service.SetArchived(callerContext(), archiveParams(archive))
}

// runCascade is runCascadeRaw for the tests that expect it to succeed, returning the writes it
// made.
func runCascade(
	t *testing.T, repo *stubVariantRepository, base *stubBaseService, archive bool,
) []dmodel.DynamicFields {
	t.Helper()

	_, err := runCascadeRaw(t, repo, base, archive)
	require.NoError(t, err)
	require.NotNil(t, repo.tranx)
	assert.True(t, repo.tranx.committed, "a successful cascade commits")
	return repo.updates
}

// Without an is_archived flag there is no direction to cascade in. The override must hand the
// request to the base service, which reports the missing required field, rather than guessing.
func TestSetArchivedWithoutFlagDelegatesToTheBase(t *testing.T) {
	base := &stubBaseService{}
	service := &ProductTemplateDomainServiceImpl{DynamicResourceService: base}

	result, err := service.SetArchived(callerContext(), dmodel.DynamicFields{
		models.ProductTemplateFieldId: testTemplateId,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, base.calls, "the base service must still decide the missing-flag error")
}

// The cascade must reach the variants that need it and leave the rest alone. The skip rule itself
// is ShouldSkipCascade; this asserts the loop honours it and writes variant ids.
func TestSetArchivedCascadesToVariantsNeedingIt(t *testing.T) {
	cascadeSource := models.ArchiveSourceTemplateCascade
	repo := &stubVariantRepository{variants: []dmodel.DynamicFields{
		archivedVariant("01VARIANTACTIVE00000000000", false, nil),
		archivedVariant("01VARIANTARCHIVED000000000", true, &cascadeSource),
	}}

	updates := runCascade(t, repo, &stubBaseService{}, true)

	require.Len(t, updates, 1, "only the unarchived variant needed the cascade")
	assert.Equal(t, "01VARIANTACTIVE00000000000", updates[0][models.ProductVariantFieldId])
	assert.Equal(t, true, updates[0][basemodel.FieldIsArchived])
	assert.Equal(t, cascadeSource.String(), updates[0][models.ProductVariantFieldArchiveSource],
		"a cascaded archive stamps its source, so a later unarchive can recognise it")
}

// The write path stamps updated_by from the context, so the scoped transaction context must carry
// the caller across. Losing it writes a null audit column and reports nothing, which is why this
// is asserted rather than assumed.
func TestCascadeWritesCarryTheCallerIdentity(t *testing.T) {
	repo := &stubVariantRepository{variants: []dmodel.DynamicFields{
		archivedVariant("01VARIANTACTIVE00000000000", false, nil),
	}}
	base := &stubBaseService{}

	runCascade(t, repo, base, true)

	require.Len(t, repo.updateContexts, 1)
	assert.Equal(t, model.Id(testCallerId), repo.updateContexts[0].GetPermissions().UserId,
		"the cascade write must know who asked for it")
	assert.NotNil(t, repo.updateContexts[0].GetDbTranx(),
		"the write must run inside the transaction")

	require.Len(t, base.seenContexts, 1)
	assert.Equal(t, model.Id(testCallerId), base.seenContexts[0].GetPermissions().UserId)
	assert.NotNil(t, base.seenContexts[0].GetDbTranx(),
		"the template write must join the same transaction")
}

// The transaction is the reason the cascade moved into the service: a failed variant write must
// not leave the template archived on its own.
func TestFailedCascadeWriteRollsBack(t *testing.T) {
	repo := &stubVariantRepository{
		variants:  []dmodel.DynamicFields{archivedVariant("01VARIANTACTIVE00000000000", false, nil)},
		updateErr: errors.New("write failed"),
	}

	_, err := runCascadeRaw(t, repo, &stubBaseService{}, true)

	require.Error(t, err)
	require.NotNil(t, repo.tranx)
	assert.False(t, repo.tranx.committed, "a failed cascade must not commit")
}

// A client error from the template write means nothing was archived, so there is nothing to
// cascade and no reason to commit.
func TestClientErrorSkipsTheCascade(t *testing.T) {
	repo := &stubVariantRepository{variants: []dmodel.DynamicFields{
		archivedVariant("01VARIANTACTIVE00000000000", false, nil),
	}}
	base := &stubBaseService{}
	base.clientErrors.Append(*ft.NewAnonymousNotFoundError())

	result, err := runCascadeRaw(t, repo, base, true)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Empty(t, repo.updates, "a template that was not archived cascades to nothing")
	assert.False(t, repo.tranx.committed)
}
