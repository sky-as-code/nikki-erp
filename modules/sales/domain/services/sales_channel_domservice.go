package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesChannelDomainServiceImpl adds the channel lifecycle to the engine's default service. It
// wraps rather than replaces: ordinary CRUD still runs through the engine's implementation.
type SalesChannelDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesChannelDomainService(base drif.DynamicResourceService) *SalesChannelDomainServiceImpl {
	return &SalesChannelDomainServiceImpl{DynamicResourceService: base}
}

// ResolveByCode answers the id and metadata of the channel an external module names by code, which
// keeps database ids out of other modules' configuration. A suspended or archived channel resolves
// but is reported as such, because the caller may be reading history rather than selling.
func (this *SalesChannelDomainServiceImpl) ResolveByCode(
	ctx corectx.Context, code string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	normalized := NormalizeChannelCode(code)
	if normalized == "" {
		return violationResult(models.SalesChannelSchemaName,
			"sales_channel.code_required",
			"resolving a sales channel requires a code"), nil
	}

	engine, err := engineFor(models.SalesChannelSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := models.FindSalesChannelByCode(ctx, engine.ResourceRepository(), normalized)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return violationResult(models.SalesChannelSchemaName,
			"sales_channel.not_found",
			"no sales channel with code '"+normalized+"'"), nil
	}
	return mutateOk(), nil
}

// Suspend stops a channel taking new business.
//
// History, returns, refunds and fiscal adjustments of existing transactions keep working: the
// status says the channel is not selling now, not that it is gone. Hence it is separate from
// is_archived, and a system channel may be suspended even though it may not be archived.
func (this *SalesChannelDomainServiceImpl) Suspend(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setStatus(ctx, channelId, models.SalesChannelStatusSuspended)
}

// Activate returns a suspended channel to service.
func (this *SalesChannelDomainServiceImpl) Activate(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setStatus(ctx, channelId, models.SalesChannelStatusActive)
}

func (this *SalesChannelDomainServiceImpl) setStatus(
	ctx corectx.Context, channelId string, status models.SalesChannelStatus,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesChannelSchemaName, func(tranxCtx corectx.Context) error {
		channel, err := loadRecord(tranxCtx, models.SalesChannelSchemaName,
			models.SalesChannelFieldId, channelId)
		if err != nil {
			return err
		}
		if channel == nil {
			result = notFoundResult(models.SalesChannelSchemaName, channelId)
			return nil
		}
		if boolOf(channel, basemodel.FieldIsArchived) {
			result = violationResult(models.SalesChannelSchemaName,
				"sales_channel.archived",
				"an archived sales channel cannot change status; unarchive it first")
			return nil
		}
		if stringOf(channel, models.SalesChannelFieldStatus) == string(status) {
			// Already there. A retry is not an error: the caller wanted this state and it holds.
			result = mutateOk()
			return nil
		}

		result = mutateOk()
		return writeChanges(tranxCtx, models.SalesChannelSchemaName, channel, dmodel.DynamicFields{
			models.SalesChannelFieldStatus: string(status),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Archive retires a channel.
//
// It refuses while any sales point under it is still unarchived, because archiving the parent would
// strand them: such a point can neither sell nor be found in the channel's list. Archive the points
// first. A system channel is never archivable: archiving the vending channel would stop every kiosk
// selling with no change to any kiosk.
func (this *SalesChannelDomainServiceImpl) Archive(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesChannelSchemaName, func(tranxCtx corectx.Context) error {
		channel, err := loadRecord(tranxCtx, models.SalesChannelSchemaName,
			models.SalesChannelFieldId, channelId)
		if err != nil {
			return err
		}
		if channel == nil {
			result = notFoundResult(models.SalesChannelSchemaName, channelId)
			return nil
		}
		if boolOf(channel, models.SalesChannelFieldIsSystem) {
			result = violationResult(models.SalesChannelSchemaName,
				"sales_channel.is_system",
				"a system sales channel cannot be archived; it is seeded and depended on by the "+
					"module that registered it")
			return nil
		}
		if boolOf(channel, basemodel.FieldIsArchived) {
			result = mutateOk()
			return nil
		}

		blocking, err := this.activeSalesPointsOf(tranxCtx, channelId)
		if err != nil {
			return err
		}
		if len(blocking) > 0 {
			result = violationResult(models.SalesChannelSchemaName,
				"sales_channel.has_active_sales_points",
				"this sales channel still has unarchived sales points; archive them first")
			return nil
		}

		result = mutateOk()
		return writeChanges(tranxCtx, models.SalesChannelSchemaName, channel, dmodel.DynamicFields{
			basemodel.FieldIsArchived: true,
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *SalesChannelDomainServiceImpl) activeSalesPointsOf(
	ctx corectx.Context, channelId string,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	return models.FindActiveSalesPointsOfChannel(
		ctx, engine.ResourceRepository(), channelId, models.MaxSalesPointsPerChannel)
}

// AssertMutable refuses a change to a channel the API may not alter, and is called by the write
// guards on plain update and delete. A system channel's name, code and owning module are fixed at
// seed time because other modules resolve against them.
func (this *SalesChannelDomainServiceImpl) AssertMutable(
	ctx corectx.Context, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	channel, err := loadRecord(ctx, models.SalesChannelSchemaName,
		models.SalesChannelFieldId, channelId)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return notFoundResult(models.SalesChannelSchemaName, channelId), nil
	}
	if boolOf(channel, models.SalesChannelFieldIsSystem) {
		return violationResult(models.SalesChannelSchemaName,
			"sales_channel.is_system",
			"a system sales channel cannot be modified or deleted through the API"), nil
	}
	return mutateOk(), nil
}
