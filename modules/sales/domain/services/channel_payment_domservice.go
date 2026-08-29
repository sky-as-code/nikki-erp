package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// ChannelPaymentDomainService owns the sales_channel_payment_rel rows.
//
// A plain service rather than a derived resource service, because a mapping is configuration of
// its channel rather than a resource a client creates and edits: no REST route, no IAM resource
// row, no CRUD. Rows are written through the repository rather than crud.ManageM2m, whose helper
// resolves ids against a locally registered schema; a payment method belongs to paymentinvoice,
// so this service validates the target through the port instead.
type ChannelPaymentDomainServiceImpl struct {
	repo drif.DynamicResourceRepository
}

func NewChannelPaymentDomainService(
	repo drif.DynamicResourceRepository,
) *ChannelPaymentDomainServiceImpl {
	return &ChannelPaymentDomainServiceImpl{repo: repo}
}

// ListMappings returns only the local half. Merging it with what paymentinvoice reports is the
// application service's job, where the port lives, and the frontend must never join two modules.
func (this *ChannelPaymentDomainServiceImpl) ListMappings(
	ctx corectx.Context, channelId string,
) ([]dmodel.DynamicFields, error) {
	return models.FindPaymentMethodsOfChannel(ctx, this.repo, channelId)
}

// FindMapping is what makes enabling idempotent and disabling repeatable: both ask first and then
// act, so a caller that retries after a lost response gets the same answer.
func (this *ChannelPaymentDomainServiceImpl) FindMapping(
	ctx corectx.Context, channelId string, paymentMethodId string,
) (dmodel.DynamicFields, error) {
	found, err := models.FindChannelPaymentMapping(ctx, this.repo, channelId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	if len(found) > 1 {
		// The composite unique makes this impossible. If it happens the constraint is gone, and
		// carrying on with the first row would quietly disable only half a mapping next call.
		return nil, errors.Errorf(
			"sales_channel_payment_rel holds %d rows for channel '%s' and method '%s'; "+
				"the (sales_channel_id, payment_method_id) unique constraint is missing",
			len(found), channelId, paymentMethodId)
	}
	return found[0], nil
}

// Enable does nothing when the mapping is already there. It does NOT validate the payment method:
// that needs the port, which lives in the application layer, and a domain service never reaches
// across a module boundary. The check and the insert need no transaction - the composite unique is
// the real guard against a concurrent double-enable.
func (this *ChannelPaymentDomainServiceImpl) Enable(
	ctx corectx.Context, channelId string, paymentMethodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	existing, err := this.FindMapping(ctx, channelId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Already enabled. Reported as success: the caller asked for a state and that state holds,
		// and refusing would make a retry after a lost response look like a failure.
		return mutateOk(), nil
	}

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "ChannelPaymentDomainService.Enable")
	}
	inserted, err := this.repo.Insert(ctx, dmodel.DynamicFields{
		models.SalesChannelPaymentRelFieldId:              string(*id),
		models.SalesChannelPaymentRelFieldSalesChannelId:  channelId,
		models.SalesChannelPaymentRelFieldPaymentMethodId: paymentMethodId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ChannelPaymentDomainService.Enable")
	}
	if inserted.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: inserted.ClientErrors}, nil
	}
	return mutateOk(), nil
}

// Disable does nothing when the mapping is not there. The row is deleted rather than flagged,
// because the row IS the state; payments already taken record which method they used and read
// nothing back through this table. A method paymentinvoice no longer reports can still be
// disabled here: removing a stale mapping is exactly the fix.
func (this *ChannelPaymentDomainServiceImpl) Disable(
	ctx corectx.Context, channelId string, paymentMethodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	existing, err := this.FindMapping(ctx, channelId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return mutateOk(), nil
	}

	deleted, err := this.repo.DeleteOne(ctx, dmodel.DynamicFields{
		models.SalesChannelPaymentRelFieldId: stringOf(
			existing, models.SalesChannelPaymentRelFieldId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "ChannelPaymentDomainService.Disable")
	}
	if deleted.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: deleted.ClientErrors}, nil
	}
	return mutateOk(), nil
}

// IsEnabled reports whether the mapping exists, not whether the method is usable: the mapping is
// Sales' configuration, usability is paymentinvoice's judgement, and a caller taking a payment
// must ask both. Default-deny: no mapping means no, never "all allowed".
func (this *ChannelPaymentDomainServiceImpl) IsEnabled(
	ctx corectx.Context, channelId string, paymentMethodId string,
) (bool, error) {
	if channelId == "" || paymentMethodId == "" {
		return false, nil
	}
	found, err := this.FindMapping(ctx, channelId, paymentMethodId)
	if err != nil {
		return false, err
	}
	return found != nil, nil
}

// DisableAllOfChannel removes every mapping a channel holds, for reconfiguring one wholesale. Not
// called when archiving: an unarchived channel must come back configured as it was.
func (this *ChannelPaymentDomainServiceImpl) DisableAllOfChannel(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	mappings, err := this.ListMappings(ctx, channelId)
	if err != nil {
		return nil, err
	}

	for _, mapping := range mappings {
		deleted, err := this.repo.DeleteOne(ctx, dmodel.DynamicFields{
			models.SalesChannelPaymentRelFieldId: stringOf(
				mapping, models.SalesChannelPaymentRelFieldId),
		})
		if err != nil {
			return nil, errors.Wrap(err, "ChannelPaymentDomainService.DisableAllOfChannel")
		}
		if deleted.ClientErrors.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: deleted.ClientErrors}, nil
		}
	}
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: len(mappings)},
		HasData: true,
	}, nil
}
