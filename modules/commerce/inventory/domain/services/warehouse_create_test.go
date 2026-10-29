package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type warehouseTestTransaction struct {
	commits   int
	rollbacks int
}

func (this *warehouseTestTransaction) Commit() error   { this.commits++; return nil }
func (this *warehouseTestTransaction) Rollback() error { this.rollbacks++; return nil }

type warehouseTestRepo struct {
	composable.CrudRepository
	tranx *warehouseTestTransaction
}

func (this *warehouseTestRepo) Schema() *dmodel.ModelSchema { return &dmodel.ModelSchema{} }

func (this *warehouseTestRepo) BeginTransaction(corectx.Context) (database.DbTransaction, error) {
	return this.tranx, nil
}

// warehouseTestBase stands in for the composable default domain service: it records whether it
// ran inside a transaction and answers with the refusal it was given.
type warehouseTestBase struct {
	composable.CrudDomainService
	repo           *warehouseTestRepo
	refusal        *ft.ClientErrors
	calls          int
	sawTransaction bool
}

func (this *warehouseTestBase) Repository() composable.CrudRepository { return this.repo }

func (this *warehouseTestBase) Create(
	ctx corectx.Context, _ composable.CreateCommand, _ ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	this.calls++
	this.sawTransaction = ctx.GetDbTranx() != nil
	return &composable.CreateResult{ClientErrors: *this.refusal}, nil
}

func newWarehouseCreateFixture() (*WarehouseDomainServiceImpl, *warehouseTestBase, *warehouseTestTransaction) {
	tranx := &warehouseTestTransaction{}
	base := &warehouseTestBase{repo: &warehouseTestRepo{tranx: tranx}, refusal: ft.NewClientErrors()}
	return NewWarehouseDomainService(base, nil), base, tranx
}

// A warehouse and its root location are created in one transaction: the default create runs inside
// it, and a refused create provisions and commits nothing.
func TestWarehouseCreateRunsTheDefaultInsideOneTransaction(t *testing.T) {
	service, base, tranx := newWarehouseCreateFixture()
	base.refusal.Append(*ft.NewBusinessViolation("code", "common:err_missing_required_field", "missing"))

	result, err := service.Create(corectx.NewRequestContext(context.Background()), dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Equal(t, 1, base.calls)
	assert.True(t, base.sawTransaction, "the default create must join the provisioning transaction")
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Zero(t, tranx.commits, "a refused create provisions and commits nothing")
	assert.Equal(t, 1, tranx.rollbacks)
}

func TestWarehouseCreateRefusesAnUnknownFlowBeforeWriting(t *testing.T) {
	service, base, _ := newWarehouseCreateFixture()

	result, err := service.Create(corectx.NewRequestContext(context.Background()), dmodel.DynamicFields{
		models.WarehouseFieldIncomingFlow: "teleport",
	})

	require.NoError(t, err)
	assert.Zero(t, base.calls)
	assert.Equal(t, 1, result.ClientErrors.Count())
}
