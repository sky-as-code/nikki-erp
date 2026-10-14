package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

func FindAwaitingOrders(ctx corectx.Context, createdAfter time.Time) ([]StaleOrder, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.OrderFieldStatus, dmodel.In,
			models.OrderStatusPending, models.OrderStatusProcessing),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldCreatedAt, dmodel.GreaterThan, createdAfter),
	)
	graph.OrderBy(basemodel.FieldCreatedAt)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  sweepPageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindAwaitingOrders")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	awaiting := make([]StaleOrder, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		order := models.NewOrderFrom(item)
		awaiting = append(awaiting, StaleOrder{
			Pk:               derefString(order.GetId()),
			OrderId:          derefString(order.GetOrderId()),
			OrderCode:        derefString(order.GetOrderCode()),
			Status:           derefString(order.GetStatus()),
			PaymentMethodId:  derefString((*string)(order.GetPaymentMethodId())),
			PaymentProfileId: derefString((*string)(order.GetPaymentProfileId())),
			ReturnUrl:        derefString(order.GetReturnUrl()),
			Metadata:         order.GetMetadata(),
		})
	}
	return awaiting, nil
}
