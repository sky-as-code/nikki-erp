package services

import (
	"context"
	"errors"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	"github.com/stretchr/testify/require"
)

type orderTotalsWriteRepo struct {
	composable.CrudRepository
	result *composable.MutateResult
	err    error
}

func (r orderTotalsWriteRepo) Update(corectx.Context, dmodel.DynamicFields) (*composable.MutateResult, error) {
	return r.result, r.err
}

func TestWriteOrderTotalsPropagatesRepositoryFailures(t *testing.T) {
	dbErr := errors.New("database update failed")
	for _, tc := range []struct {
		name   string
		result *composable.MutateResult
		err    error
		want   string
	}{
		{name: "client error with nil Go error", result: &composable.MutateResult{
			ClientErrors: ft.ClientErrors{*dmodel.NewInvalidDataTypeErr("tax_snapshot", "jsonb")},
		}, want: "tax_snapshot"},
		{name: "database error", err: dbErr, want: dbErr.Error()},
		{name: "missing result", want: "order totals update returned no result"},
		{name: "success", result: &composable.MutateResult{HasData: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := repoFor
			t.Cleanup(func() { repoFor = previous })
			repoFor = func(string) (composable.CrudRepository, error) {
				return orderTotalsWriteRepo{result: tc.result, err: tc.err}, nil
			}
			err := writeOrderTotals(corectx.NewRequestContext(context.Background()), "order", nil, pricing.Result{}, nil)
			if tc.want == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.want)
				if tc.err != nil {
					require.ErrorIs(t, err, tc.err)
				}
			}
		})
	}
}
