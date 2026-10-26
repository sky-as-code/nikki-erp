package services

import (
	"context"
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

// NewConsumptionDomainService is handed the composable default by the itstorage_consumption onion.
func NewConsumptionDomainService(base composable.CrudDomainService) itConsumption.ConsumptionDomainService {
	return &ConsumptionDomainServiceImpl{
		CrudDomainService: base,
		repo:              base.Repository().(itConsumption.ConsumptionRepository),
	}
}

// ConsumptionDomainServiceImpl reserves an upload's bytes before it streams and settles them once
// its record commits or the upload fails. The reservation is what makes the quota check atomic:
// reading the used bytes and comparing them would let concurrent uploads pass together.
type ConsumptionDomainServiceImpl struct {
	composable.CrudDomainService
	repo itConsumption.ConsumptionRepository
}

// errCapExceeded rolls the reservation back through the transaction and never leaves Reserve.
var errCapExceeded = errors.New("the storage quota would be exceeded")

// Reserve takes the capped owner first, so a refused upload touches no other row. Without a
// caller's transaction it runs in a short one of its own, committed before the content streams;
// inside one (a duplicate during a record's save) it joins it, and settles or rolls back with it.
func (this *ConsumptionDomainServiceImpl) Reserve(
	ctx corectx.Context, reservation itConsumption.Reservation,
) (bool, error) {
	if len(reservation.Owners) == 0 {
		return true, nil
	}
	if err := this.repo.Ensure(ctx, reservation.Owners); err != nil {
		return false, err
	}
	err := composable.InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		for index, owner := range reservation.Owners {
			capBytes := reservation.CapBytes
			if index > 0 {
				capBytes = nil
			}
			reserved, err := this.repo.Reserve(tranxCtx, owner, reservation.Bytes, capBytes)
			if err != nil {
				return err
			}
			if !reserved {
				return errCapExceeded
			}
		}
		return nil
	})
	if stdErr.Is(err, errCapExceeded) {
		return false, nil
	}
	return err == nil, errors.Wrap(err, "reserve storage consumption")
}

func (this *ConsumptionDomainServiceImpl) Commit(
	ctx corectx.Context, reservation itConsumption.Reservation,
) error {
	return this.repo.Commit(ctx, reservation.Owners, reservation.Bytes)
}

// Release survives the request: it is called on failure paths, where the request context may
// already be cancelled.
func (this *ConsumptionDomainServiceImpl) Release(ctx corectx.Context, reservation itConsumption.Reservation) {
	if len(reservation.Owners) == 0 || reservation.Bytes == 0 {
		return
	}
	// A reservation made inside the caller's transaction rolls back with it. Releasing it on
	// another connection would also wait on the row lock that very transaction holds.
	if ctx.GetDbTranx() != nil {
		return
	}
	detached := detachedContext(ctx)
	if err := this.repo.Release(detached, reservation.Owners, reservation.Bytes); err != nil {
		ctx.GetLogger().Warn("itstorage: failed to release reserved bytes; they stay reserved until reconciled",
			logging.Attr{"bytes": reservation.Bytes, "error": err.Error()})
	}
}

func (this *ConsumptionDomainServiceImpl) Subtract(
	ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64,
) error {
	if bytes == 0 {
		return nil
	}
	return this.repo.Subtract(ctx, owners, bytes)
}

func (this *ConsumptionDomainServiceImpl) SubtractEach(
	ctx corectx.Context, ownerType string, bytesByOwner map[model.Id]int64,
) error {
	return this.repo.SubtractEach(ctx, ownerType, bytesByOwner)
}

func (this *ConsumptionDomainServiceImpl) DeleteOwner(ctx corectx.Context, owner models.ConsumptionOwner) error {
	return this.repo.DeleteOwner(ctx, owner)
}

// detachedContext keeps the request's identity and tenant but neither its cancellation nor its
// transaction.
func detachedContext(ctx corectx.Context) corectx.Context {
	detached := corectx.CloneWithInner(ctx, context.WithoutCancel(ctx.InnerContext()))
	detached.SetDbTranx(nil)
	return detached
}
