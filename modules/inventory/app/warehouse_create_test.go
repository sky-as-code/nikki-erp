package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
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

type warehouseTestDomain struct {
	composable.CrudDomainService
	repo *warehouseTestRepo
}

func (this *warehouseTestDomain) Repository() composable.CrudRepository { return this.repo }

// warehouseTestBase stands in for the composable default: it records whether it ran inside a
// transaction and answers with the refusal it was given.
type warehouseTestBase struct {
	composable.CrudApplicationService
	refusal        *ft.ClientErrors
	calls          int
	sawTransaction bool
}

func (this *warehouseTestBase) Create(
	ctx corectx.Context, _ composable.CreateCommand,
) (*composable.CreateResult, error) {
	this.calls++
	this.sawTransaction = ctx.GetDbTranx() != nil
	return &composable.CreateResult{ClientErrors: *this.refusal}, nil
}

func newWarehouseCreateFixture() (*WarehouseApplicationServiceImpl, *warehouseTestBase, *warehouseTestTransaction) {
	tranx := &warehouseTestTransaction{}
	domain := &warehouseTestDomain{repo: &warehouseTestRepo{tranx: tranx}}
	warehouseSvc := &services.WarehouseDomainServiceImpl{CrudDomainService: domain}
	base := &warehouseTestBase{refusal: ft.NewClientErrors()}
	orchestrator := &WarehouseAppServiceImpl{warehouseSvc: warehouseSvc}
	return &WarehouseApplicationServiceImpl{
		CrudApplicationService: base,
		WarehouseAppService:    orchestrator,
		orchestrator:           orchestrator,
		warehouseSvc:           warehouseSvc,
	}, base, tranx
}

// The REST create reaches the onion's Create, so that is where a warehouse must get its system
// locations; the default create runs inside the transaction they are provisioned in.
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
