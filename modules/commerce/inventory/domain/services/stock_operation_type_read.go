package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

var _ itStock.StockOperationTypeReadService = (*StockTransferDomainServiceImpl)(nil)

// FindOperationTypeByCode resolves one operation type by the pair that identifies it: `code` is
// unique within an organization, so at most one row matches.
func (this *StockTransferDomainServiceImpl) FindOperationTypeByCode(
	ctx corectx.Context, query itStock.FindOperationTypeQuery,
) (*itStock.FindOperationTypeResult, error) {
	if query.OrgId == "" || query.Code == "" {
		return &itStock.FindOperationTypeResult{}, nil
	}

	engine, err := repoFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockOperationTypeFieldOrgId, dmodel.Equals, query.OrgId),
		*dmodel.NewSearchNode().NewCondition(models.StockOperationTypeFieldCode, dmodel.Equals, query.Code),
	)

	found, err := engine.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindOperationTypeByCode")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return &itStock.FindOperationTypeResult{}, nil
	}

	operationType := models.NewStockOperationTypeFrom(found.Data.Items[0])
	return &itStock.FindOperationTypeResult{
		Data: itStock.FindOperationTypeResultData{
			OperationType: itStock.OperationTypeRef{
				Id:            derefString(operationType.GetId()),
				OperationCode: derefString(operationType.GetOperationCode()),
				IsArchived:    isArchived(operationType.GetFieldData()),
			},
		},
		HasData: true,
	}, nil
}
