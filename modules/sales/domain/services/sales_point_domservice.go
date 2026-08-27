package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesPointDomainServiceImpl adds the sales point lifecycle to the engine's default service.
type SalesPointDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesPointDomainService(base drif.DynamicResourceService) *SalesPointDomainServiceImpl {
	return &SalesPointDomainServiceImpl{DynamicResourceService: base}
}

// Suspend stops a sales point taking new orders.
//
// Returns, refunds and history keep working, which is what a temporarily offline kiosk needs: the
// machine stops selling, but the money already taken through it stays refundable.
func (this *SalesPointDomainServiceImpl) Suspend(
	ctx corectx.Context, salesPointId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setStatus(ctx, salesPointId, models.SalesPointStatusSuspended)
}

// Activate returns a suspended sales point to service.
//
// It refuses an archived point. Activating one would resurrect a decommissioned kiosk through an
// operation whose name suggests nothing of the sort; bringing one back is Unarchive, deliberately
// a separate power so a role may resume selling without being able to revive retired machines.
func (this *SalesPointDomainServiceImpl) Activate(
	ctx corectx.Context, salesPointId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setStatus(ctx, salesPointId, models.SalesPointStatusActive)
}

func (this *SalesPointDomainServiceImpl) setStatus(
	ctx corectx.Context, salesPointId string, status models.SalesPointStatus,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesPointSchemaName, func(tranxCtx corectx.Context) error {
		point, err := loadRecord(tranxCtx, models.SalesPointSchemaName,
			models.SalesPointFieldId, salesPointId)
		if err != nil {
			return err
		}
		if point == nil {
			result = notFoundResult(models.SalesPointSchemaName, salesPointId)
			return nil
		}
		if boolOf(point, basemodel.FieldIsArchived) {
			result = violationResult(models.SalesPointSchemaName,
				"sales_point.archived",
				"an archived sales point cannot change status; unarchive it first")
			return nil
		}
		if stringOf(point, models.SalesPointFieldStatus) == string(status) {
			result = mutateOk()
			return nil
		}

		result = mutateOk()
		return writeChanges(tranxCtx, models.SalesPointSchemaName, point, dmodel.DynamicFields{
			models.SalesPointFieldStatus: string(status),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Archive retires a sales point. Idempotent: archiving an archived point reports success.
func (this *SalesPointDomainServiceImpl) Archive(
	ctx corectx.Context, salesPointId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setArchived(ctx, salesPointId, true)
}

// Unarchive brings a retired sales point back.
//
// It refuses while the parent channel is archived, because the result would be a live point under a
// dead channel — reachable by id, absent from every listing that starts at the channel.
func (this *SalesPointDomainServiceImpl) Unarchive(
	ctx corectx.Context, salesPointId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setArchived(ctx, salesPointId, false)
}

func (this *SalesPointDomainServiceImpl) setArchived(
	ctx corectx.Context, salesPointId string, archived bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesPointSchemaName, func(tranxCtx corectx.Context) error {
		point, err := loadRecord(tranxCtx, models.SalesPointSchemaName,
			models.SalesPointFieldId, salesPointId)
		if err != nil {
			return err
		}
		if point == nil {
			result = notFoundResult(models.SalesPointSchemaName, salesPointId)
			return nil
		}
		if boolOf(point, basemodel.FieldIsArchived) == archived {
			result = mutateOk()
			return nil
		}

		if !archived {
			channelId := stringOf(point, models.SalesPointFieldSalesChannelId)
			channel, err := loadRecord(tranxCtx, models.SalesChannelSchemaName,
				models.SalesChannelFieldId, channelId)
			if err != nil {
				return err
			}
			if channel != nil && boolOf(channel, basemodel.FieldIsArchived) {
				result = violationResult(models.SalesPointSchemaName,
					"sales_point.channel_archived",
					"this sales point belongs to an archived sales channel; unarchive the channel first")
				return nil
			}
		}

		result = mutateOk()
		return writeChanges(tranxCtx, models.SalesPointSchemaName, point, dmodel.DynamicFields{
			basemodel.FieldIsArchived: archived,
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// AssertCreatable refuses a sales point whose channel cannot take one.
//
// Called by the create guard. The channel must exist, be active and be unarchived: a point created
// under a suspended channel could never sell, and one under an archived channel would be invisible
// from the moment it was written.
func (this *SalesPointDomainServiceImpl) AssertCreatable(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if channelId == "" {
		return violationResult(models.SalesPointSchemaName,
			"sales_point.channel_required",
			"a sales point must name the sales channel it belongs to"), nil
	}

	channel, err := loadRecord(ctx, models.SalesChannelSchemaName,
		models.SalesChannelFieldId, channelId)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return violationResult(models.SalesPointSchemaName,
			"sales_channel.not_found",
			"no sales channel with id '"+channelId+"'"), nil
	}
	if boolOf(channel, basemodel.FieldIsArchived) {
		return violationResult(models.SalesPointSchemaName,
			"sales_channel.archived",
			"an archived sales channel cannot take new sales points"), nil
	}
	if stringOf(channel, models.SalesChannelFieldStatus) != string(models.SalesChannelStatusActive) {
		return violationResult(models.SalesPointSchemaName,
			"sales_channel.suspended",
			"a suspended sales channel cannot take new sales points"), nil
	}
	return mutateOk(), nil
}

// FindByExternalReference resolves the point a module already registered for one of its records.
//
// This is what makes registration idempotent: a caller retrying after a timeout is handed the point
// it created the first time rather than making a second one. Returning "not found" as a nil record
// rather than a refusal is deliberate — the caller's next step is to create, and an error here
// would make the ordinary first-time path look like a failure.
func (this *SalesPointDomainServiceImpl) FindByExternalReference(
	ctx corectx.Context, channelId string, externalReferenceId string,
) (dmodel.DynamicFields, error) {
	if channelId == "" || externalReferenceId == "" {
		return nil, nil
	}
	engine, err := engineFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := models.FindSalesPointByExternalReferenceId(
		ctx, engine.ResourceRepository(), channelId, externalReferenceId)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return found[0], nil
}

// DeleteOrArchive removes a sales point, or retires it when history forbids removal.
//
// One operation rather than two because the choice belongs where the data is. A caller cannot know
// whether a point has ever carried an order without asking, and making it ask, branch, and call
// again turns one decision into a race: an order could arrive between the question and the delete.
//
// The result reports which happened, so a client can tell the operator "removed" from "archived,
// because it has sales history".
func (this *SalesPointDomainServiceImpl) DeleteOrArchive(
	ctx corectx.Context, salesPointId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesPointSchemaName, func(tranxCtx corectx.Context) error {
		point, err := loadRecord(tranxCtx, models.SalesPointSchemaName,
			models.SalesPointFieldId, salesPointId)
		if err != nil {
			return err
		}
		if point == nil {
			result = notFoundResult(models.SalesPointSchemaName, salesPointId)
			return nil
		}

		// Until sales orders exist there is nothing that can reference a point, so the safe branch
		// is unreachable and the point is always removable. SALES-007 adds the order table; when it
		// does, this is the one place that has to learn to count references, and the operation's
		// contract does not change.
		hasHistory, err := this.hasSalesHistory(tranxCtx, salesPointId)
		if err != nil {
			return err
		}
		if hasHistory {
			if boolOf(point, basemodel.FieldIsArchived) {
				result = mutateOk()
				return nil
			}
			result = mutateOk()
			return writeChanges(tranxCtx, models.SalesPointSchemaName, point, dmodel.DynamicFields{
				basemodel.FieldIsArchived: true,
			})
		}

		engine, err := engineFor(models.SalesPointSchemaName)
		if err != nil {
			return err
		}
		_, err = engine.ResourceRepository().DeleteOne(tranxCtx, dmodel.DynamicFields{
			models.SalesPointFieldId: salesPointId,
		})
		if err != nil {
			return err
		}
		result = mutateOk()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// hasSalesHistory reports whether anything references this sales point.
//
// Always false for now: the sales order resource does not exist yet, so nothing can reference a
// point. It is a named function rather than an inline false so that the rule has one home when
// SALES-007 lands, and so the reason is written down where somebody would look for it.
func (this *SalesPointDomainServiceImpl) hasSalesHistory(
	ctx corectx.Context, salesPointId string,
) (bool, error) {
	return false, nil
}
