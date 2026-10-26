package services

import (
	"context"
	"errors"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/stretchr/testify/require"
)

type draftInsertProbe struct {
	composable.CrudRepository
	fields dmodel.DynamicFields
	stop   error
}

func (r *draftInsertProbe) Insert(_ corectx.Context, fields dmodel.DynamicFields) (*dyn.OpResult[int], error) {
	r.fields = fields
	return nil, r.stop
}

func (r *draftInsertProbe) BeginTransaction(corectx.Context) (database.DbTransaction, error) {
	return draftInsertTransaction{}, nil
}

type draftInsertTransaction struct{}

func (draftInsertTransaction) Commit() error   { return nil }
func (draftInsertTransaction) Rollback() error { return nil }

// Stop at Insert so these tests inspect the actual maps without running pricing,
// stage events or a real database transaction.
func TestDraftOrderInsertsExplicitArchiveDefault(t *testing.T) {
	for _, schema := range []string{models.SalesOrderSchemaName, models.SalesOrderLineSchemaName, models.SalesOrderLineAllocationSchemaName} {
		t.Run(schema, func(t *testing.T) {
			stop := errors.New("insert captured")
			probe := &draftInsertProbe{stop: stop}
			previous := repoFor
			t.Cleanup(func() { repoFor = previous })
			repoFor = func(name string) (composable.CrudRepository, error) {
				require.Equal(t, schema, name)
				return probe, nil
			}
			ctx := corectx.NewRequestContext(context.Background())
			var err error
			switch schema {
			case models.SalesOrderSchemaName:
				_, _, err = writeDraftOrder(ctx, CreateOrderParams{SalesPointId: "point"}, dmodel.DynamicFields{basemodel.FieldOrgId: "org"}, "channel")
			case models.SalesOrderLineSchemaName:
				_, err = writeOrderLines(ctx, "order", "org", []CreateOrderLine{{ProductVariantId: "variant"}})
			case models.SalesOrderLineAllocationSchemaName:
				err = writeOrderLineAllocations(ctx, []string{"line"}, "org", []CreateOrderLine{{Allocations: []CreateOrderLineAllocation{{LocationId: "location"}}}})
			}
			require.ErrorIs(t, err, stop)
			require.Contains(t, probe.fields, basemodel.FieldIsArchived)
			require.Equal(t, false, probe.fields[basemodel.FieldIsArchived])
			require.Equal(t, "org", probe.fields[basemodel.FieldOrgId])
		})
	}
}
