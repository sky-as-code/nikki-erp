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
// It is a plain service rather than a derived resource service, because the junction has no engine
// of its own: a mapping is not a resource a client creates, reads and edits — it is configuration of
// its channel, reached only through the channel's own capabilities. That is the same split
// vending_machine_new makes for vdmc_kiosk_payment_rel, and it is why there is no REST route, no
// IAM resource row and no CRUD for these rows.
//
// The rows are written through the repository directly rather than through crud.ManageM2m. The
// helper resolves the associated ids against a locally registered destination schema, and a payment
// method is not one: it belongs to paymentinvoice. Validating the target is therefore this
// service's job, and it does it through the port rather than by reading another module's table.
type ChannelPaymentDomainServiceImpl struct {
	repo drif.DynamicResourceRepository
}

func NewChannelPaymentDomainService(
	repo drif.DynamicResourceRepository,
) *ChannelPaymentDomainServiceImpl {
	return &ChannelPaymentDomainServiceImpl{repo: repo}
}

// ListMappings answers which payment methods a channel is configured to accept.
//
// It returns only the local half. Merging it with what paymentinvoice reports is the application
// service's job, because that is where the port lives — and merging is the whole point: the
// frontend must never join across two modules (CR §29).
func (this *ChannelPaymentDomainServiceImpl) ListMappings(
	ctx corectx.Context, channelId string,
) ([]dmodel.DynamicFields, error) {
	return models.FindPaymentMethodsOfChannel(ctx, this.repo, channelId)
}

// FindMapping resolves the one row mapping a channel to a method, or nil.
//
// This is what makes enabling idempotent and disabling repeatable: both operations ask first and
// then act, so a caller that retries after a lost response gets the same answer as the call that
// succeeded.
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
		// carrying on with the first row would hide that while quietly disabling only half a
		// mapping on the next call.
		return nil, errors.Errorf(
			"sales_channel_payment_rel holds %d rows for channel '%s' and method '%s'; "+
				"the (sales_channel_id, payment_method_id) unique constraint is missing",
			len(found), channelId, paymentMethodId)
	}
	return found[0], nil
}

// Enable writes the mapping, and does nothing when it is already there (CR §31).
//
// It does NOT validate the payment method: that check needs the port, which lives in the
// application layer, and a domain service never reaches across a module boundary. The caller
// validates first and then calls this — see EnableChannelPaymentMethod.
//
// The existence check and the insert are not in one transaction, and do not need to be: the
// composite unique is the real guard against a concurrent double-enable, and losing that race means
// the insert fails on a constraint that says exactly what happened. A transaction here would only
// move the same collision to commit time.
func (this *ChannelPaymentDomainServiceImpl) Enable(
	ctx corectx.Context, channelId string, paymentMethodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	existing, err := this.FindMapping(ctx, channelId, paymentMethodId)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Already enabled. Reported as success, not as a violation: the caller asked for a state
		// and that state holds. Refusing here would make a retry after a lost response look like a
		// failure and drive a caller to keep retrying.
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

// Disable removes the mapping, and does nothing when it is not there (CR §32).
//
// The row is deleted rather than flagged, because the row IS the state (CR §27). Deleting it does
// not touch payments already taken through that method: a payment records which method it used, and
// nothing about it reads back through this table.
//
// A method that paymentinvoice no longer reports can still be disabled through here, which is
// deliberate — CR §34 requires that a stale mapping stay removable, since removing it is exactly
// the fix an administrator is trying to apply.
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

// IsEnabled answers whether a channel accepts one payment method.
//
// This is the query SALES-027 calls at payment time, and it is deliberately narrow: it reports
// whether the mapping exists, not whether the method is usable. Those are two different questions
// with two different owners — the mapping is Sales' configuration, usability is paymentinvoice's
// judgement — and a caller taking a payment must ask both.
//
// Default-deny (CR §76): no mapping means no, never "all allowed". A channel that has never been
// configured accepts nothing, which is the correct default for something that moves money.
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

// DisableAllOfChannel removes every mapping a channel holds.
//
// Archiving a channel leaves its mappings behind on purpose, so this is not called from there: a
// channel that is unarchived must come back configured as it was, and silently dropping its payment
// configuration on archive would make that impossible to restore. It exists for the case where a
// channel is being reconfigured wholesale.
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
