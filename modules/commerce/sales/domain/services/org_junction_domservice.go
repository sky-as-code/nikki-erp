package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The association rows of an organization's sales configuration, as the org lifecycle writes and
// removes them. Unlike Enable, which serves an administrator editing one mapping, these carry the
// org_id every association row belongs to.

// EnableInOrg maps a payment method onto a channel of the organization.
func (this *ChannelPaymentDomainServiceImpl) EnableInOrg(
	ctx corectx.Context, orgId model.Id, channelId string, paymentMethodId string,
) error {
	return insertJunction(ctx, this.repo, dmodel.DynamicFields{
		basemodel.FieldOrgId:                              string(orgId),
		models.SalesChannelPaymentRelFieldSalesChannelId:  channelId,
		models.SalesChannelPaymentRelFieldPaymentMethodId: paymentMethodId,
	})
}

func (this *ChannelPaymentDomainServiceImpl) DeleteByOrg(ctx corectx.Context, orgId model.Id) (int, error) {
	return this.repo.DeleteByOrg(ctx, orgId)
}

func (this *PointPaymentDomainServiceImpl) DeleteByOrg(ctx corectx.Context, orgId model.Id) (int, error) {
	return this.repo.DeleteByOrg(ctx, orgId)
}

// AllowFulfillmentMethodOnChannel permits a fulfillment method on a channel of the organization;
// IsAllowedOnChannel reads the same rows.
func AllowFulfillmentMethodOnChannel(
	ctx corectx.Context, orgId model.Id, channelId string, methodId string,
) error {
	repo, err := repoFor(models.SalesChannelFulfillmentMethodSchemaName)
	if err != nil {
		return err
	}
	return insertJunction(ctx, repo, dmodel.DynamicFields{
		basemodel.FieldOrgId: string(orgId),
		models.SalesChannelFulfillmentMethodFieldSalesChannelId:      channelId,
		models.SalesChannelFulfillmentMethodFieldFulfillmentMethodId: methodId,
	})
}

func DeleteOrgChannelFulfillmentMethods(ctx corectx.Context, orgId model.Id) error {
	repo, err := repoFor(models.SalesChannelFulfillmentMethodSchemaName)
	if err != nil {
		return err
	}
	_, err = repo.DeleteByOrg(ctx, orgId)
	return err
}

func insertJunction(ctx corectx.Context, repo composable.CrudRepository, fields dmodel.DynamicFields) error {
	id, err := model.NewId()
	if err != nil {
		return err
	}
	fields[basemodel.FieldId] = string(*id)
	inserted, err := repo.Insert(ctx, fields)
	if err != nil {
		return errors.Wrapf(err, "insert %s", repo.Schema().Name())
	}
	if inserted.ClientErrors.Count() > 0 {
		return errors.Wrapf(inserted.ClientErrors.ToError(), "insert %s", repo.Schema().Name())
	}
	return nil
}
