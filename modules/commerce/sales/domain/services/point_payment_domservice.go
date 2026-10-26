package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

type PointPaymentDomainServiceImpl struct {
	repo composable.CrudRepository
}

func NewPointPaymentDomainService(
	repo composable.CrudRepository,
) *PointPaymentDomainServiceImpl {
	return &PointPaymentDomainServiceImpl{repo: repo}
}

func (this *PointPaymentDomainServiceImpl) ListMappings(
	ctx corectx.Context, salesPointId string,
) ([]dmodel.DynamicFields, error) {
	return models.FindPaymentMethodsOfPoint(ctx, this.repo, salesPointId)
}

func (this *PointPaymentDomainServiceImpl) FindMapping(
	ctx corectx.Context, salesPointId string, paymentMethodId string,
) (dmodel.DynamicFields, error) {
	found, err := models.FindPointPaymentMapping(ctx, this.repo, salesPointId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	if len(found) > 1 {
		return nil, errors.Errorf(
			"sales_point_payment_rel holds %d rows for point '%s' and method '%s'; "+
				"the (sales_point_id, payment_method_id) unique constraint is missing",
			len(found), salesPointId, paymentMethodId)
	}
	return found[0], nil
}

func (this *PointPaymentDomainServiceImpl) Enable(
	ctx corectx.Context, salesPointId string, paymentMethodId string, paymentProfileId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	existing, err := this.FindMapping(ctx, salesPointId, paymentMethodId)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		current := stringOf(existing, models.SalesPointPaymentRelFieldPaymentProfileId)
		if current == paymentProfileId {
			return mutateOk(), nil
		}
		updated, err := this.repo.Update(ctx, dmodel.DynamicFields{
			models.SalesPointPaymentRelFieldId: stringOf(
				existing, models.SalesPointPaymentRelFieldId),
			models.SalesPointPaymentRelFieldPaymentProfileId: nullableId(paymentProfileId),
		})
		if err != nil {
			return nil, errors.Wrap(err, "PointPaymentDomainService.Enable")
		}
		if updated.ClientErrors.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: updated.ClientErrors}, nil
		}
		return mutateOk(), nil
	}

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "PointPaymentDomainService.Enable")
	}
	inserted, err := this.repo.Insert(ctx, dmodel.DynamicFields{
		models.SalesPointPaymentRelFieldId:               string(*id),
		models.SalesPointPaymentRelFieldSalesPointId:     salesPointId,
		models.SalesPointPaymentRelFieldPaymentMethodId:  paymentMethodId,
		models.SalesPointPaymentRelFieldPaymentProfileId: nullableId(paymentProfileId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "PointPaymentDomainService.Enable")
	}
	if inserted.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: inserted.ClientErrors}, nil
	}
	return mutateOk(), nil
}

func (this *PointPaymentDomainServiceImpl) Disable(
	ctx corectx.Context, salesPointId string, paymentMethodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	existing, err := this.FindMapping(ctx, salesPointId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return mutateOk(), nil
	}

	deleted, err := this.repo.DeleteOne(ctx, dmodel.DynamicFields{
		models.SalesPointPaymentRelFieldId: stringOf(
			existing, models.SalesPointPaymentRelFieldId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "PointPaymentDomainService.Disable")
	}
	if deleted.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: deleted.ClientErrors}, nil
	}
	return mutateOk(), nil
}

func (this *PointPaymentDomainServiceImpl) IsEnabled(
	ctx corectx.Context, salesPointId string, paymentMethodId string,
) (bool, error) {
	if salesPointId == "" || paymentMethodId == "" {
		return false, nil
	}
	found, err := this.FindMapping(ctx, salesPointId, paymentMethodId)
	if err != nil {
		return false, err
	}
	return found != nil, nil
}
