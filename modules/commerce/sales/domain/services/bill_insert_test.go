package services

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/stretchr/testify/require"
)

type billStore struct {
	bills       map[string]dmodel.DynamicFields
	lines       []dmodel.DynamicFields
	sourceLines []dmodel.DynamicFields
	failSchema  string
	calls       []string
}
type billRepo struct {
	composable.CrudRepository
	store  *billStore
	schema string
}

func (r *billRepo) Search(corectx.Context, dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{HasData: true, Data: dyn.PagedResultData[dmodel.DynamicFields]{Items: r.store.sourceLines}}, nil
}
func (r *billRepo) Insert(_ corectx.Context, row dmodel.DynamicFields) (*dyn.OpResult[int], error) {
	s := r.store
	s.calls = append(s.calls, r.schema)
	if r.schema == s.failSchema || (r.schema == models.SalesBillLineSchemaName && s.bills[stringOf(row, models.SalesBillLineFieldSalesBillId)] == nil) {
		return &dyn.OpResult[int]{ClientErrors: ft.ClientErrors{*ft.NewBusinessViolation("sales_bill_id", "foreign_key_violation", "referenced bill does not exist")}}, nil
	}
	if r.schema == models.SalesBillSchemaName {
		s.bills[stringOf(row, "id")] = row
	} else {
		s.lines = append(s.lines, row)
	}
	return &dyn.OpResult[int]{HasData: true, Data: 1}, nil
}
func installBillStore(t *testing.T, s *billStore) {
	t.Helper()
	previous := repoFor
	t.Cleanup(func() { repoFor = previous })
	repoFor = func(schema string) (composable.CrudRepository, error) {
		return &billRepo{store: s, schema: schema}, nil
	}
}
func TestBillParentsExistBeforeAllocations(t *testing.T) {
	for _, mode := range []string{"initial", "split", "merge"} {
		t.Run(mode, func(t *testing.T) {
			s := &billStore{bills: map[string]dmodel.DynamicFields{}}
			installBillStore(t, s)
			ctx := corectx.NewRequestContext(context.Background())
			source := dmodel.DynamicFields{"id": "source", "sales_order_id": "order", "bill_number": "BILL-source", "currency_code": "VND", "org_id": "org", "subtotal": "100", "tax_total": "10", "grand_total": "110"}
			allocation := dmodel.DynamicFields{"sales_order_line_id": "L1", "quantity": "2", "allocated_net_amount": "100", "allocated_tax_amount": "10", "allocated_total_amount": "110"}
			var err error
			switch mode {
			case "initial":
				s.sourceLines = []dmodel.DynamicFields{orderLineRecord("L1", "110", "10", "2"), orderLineRecord("FREE", "0", "0", "1")}
				_, err = CreateInitialBill(ctx, source, SalesPolicy{RoundingScale: 4})
				require.Len(t, s.lines, 2)
			case "split":
				_, _, err = writeSplitParts(ctx, source, []dmodel.DynamicFields{allocation}, SplitBillParams{Parts: []SplitBillPart{{Allocations: map[string]decimal.Decimal{"L1": decimal.NewFromInt(1)}}, {Allocations: map[string]decimal.Decimal{"L1": decimal.NewFromInt(1)}}}}, SalesPolicy{RoundingScale: 4})
				require.Len(t, s.bills, 2)
			case "merge":
				s.sourceLines = []dmodel.DynamicFields{allocation}
				_, err = writeMergedBill(ctx, "target", "order", []dmodel.DynamicFields{source})
			}
			require.NoError(t, err)
			sum := decimal.Zero
			for id, bill := range s.bills {
				require.Contains(t, bill, "is_archived", "sales_bills requires an explicit archive default")
				require.Equal(t, false, bill["is_archived"])
				net, tax, total := decimal.Zero, decimal.Zero, decimal.Zero
				for _, line := range s.lines {
					if stringOf(line, "sales_bill_id") == id {
						net = net.Add(decimalOf(line, "allocated_net_amount"))
						tax = tax.Add(decimalOf(line, "allocated_tax_amount"))
						total = total.Add(decimalOf(line, "allocated_total_amount"))
					}
				}
				require.True(t, decimalOf(bill, "subtotal").Equal(net))
				require.True(t, decimalOf(bill, "tax_total").Equal(tax))
				require.True(t, decimalOf(bill, "total_amount").Equal(total))
				sum = sum.Add(total)
			}
			require.True(t, sum.Equal(decimal.NewFromInt(110)))
		})
	}
}
func TestInitialBillStopsOnRepositoryClientErrors(t *testing.T) {
	for _, schema := range []string{models.SalesBillSchemaName, models.SalesBillLineSchemaName} {
		t.Run(schema, func(t *testing.T) {
			s := &billStore{bills: map[string]dmodel.DynamicFields{}, failSchema: schema, sourceLines: []dmodel.DynamicFields{orderLineRecord("L1", "1", "0", "1"), orderLineRecord("L2", "1", "0", "1")}}
			installBillStore(t, s)
			id, err := CreateInitialBill(corectx.NewRequestContext(context.Background()), dmodel.DynamicFields{"id": "order", "subtotal": "2", "grand_total": "2"}, SalesPolicy{RoundingScale: 4})
			require.Empty(t, id)
			require.ErrorContains(t, err, "referenced bill does not exist")
			if schema == models.SalesBillSchemaName {
				require.Equal(t, []string{schema}, s.calls)
			} else {
				require.Equal(t, []string{models.SalesBillSchemaName, schema}, s.calls)
			}
		})
	}
}
