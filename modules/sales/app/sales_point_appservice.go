package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	it "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
)

type SalesPointApplicationServiceImpl struct{}

func NewSalesPointApplicationServiceImpl() it.SalesPointAppService {
	return &SalesPointApplicationServiceImpl{}
}

// CreateSalesPoint registers one selling place under a channel, idempotent on (channel, external
// reference id) so a lost response cannot leave a kiosk mapped to two points with split orders. The
// channel is named by code, never by id, so nothing outside Sales stores a Sales identifier.
func (this *SalesPointApplicationServiceImpl) CreateSalesPoint(
	ctx corectx.Context, command it.CreateSalesPointCommand,
) (*it.CreateSalesPointResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionCreate,
		c.SalesPointResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.CreateSalesPointResult{ClientErrors: *cErrs}, nil
	}
	if command.Name == "" {
		return pointRejection("sales_point.name_required",
			"a sales point requires a name"), nil
	}

	channel, err := this.resolveChannel(ctx, command.SalesChannelCode)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return pointRejection("sales_channel.not_found",
			"no sales channel with code '"+
				services.NormalizeChannelCode(command.SalesChannelCode)+"'"), nil
	}

	channelId := stringOf(channel, models.SalesChannelFieldId)
	pointEngine, err := services.EngineFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	service, ok := pointEngine.ResourceService().(*services.SalesPointDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales point engine is not running the derived sales point service; " +
				"SalesModule.Init must install it before CreateSalesPoint runs")
	}

	// The retry path, checked before channel usability: re-registering an existing point must
	// succeed even if the channel has since been suspended.
	if command.ExternalReferenceId != "" {
		existing, err := service.FindByExternalReference(ctx, channelId, command.ExternalReferenceId)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return &it.CreateSalesPointResult{
				HasData: true,
				Data: it.SalesPointData{
					Id:               stringOf(existing, models.SalesPointFieldId),
					SalesChannelId:   channelId,
					SalesChannelCode: stringOf(channel, models.SalesChannelFieldCode),
					AlreadyExisted:   true,
				},
			}, nil
		}
	}

	creatable, err := service.AssertCreatable(ctx, channelId)
	if err != nil {
		return nil, err
	}
	if creatable.ClientErrors.Count() > 0 {
		return &it.CreateSalesPointResult{ClientErrors: creatable.ClientErrors}, nil
	}

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "CreateSalesPoint")
	}
	fields := dmodel.DynamicFields{
		models.SalesPointFieldId:             string(*id),
		models.SalesPointFieldSalesChannelId: channelId,
		models.SalesPointFieldName:           command.Name,
		models.SalesPointFieldStatus:         string(models.SalesPointStatusActive),
	}
	if command.Code != "" {
		fields[models.SalesPointFieldCode] = command.Code
	}
	if command.ExternalReferenceId != "" {
		fields[models.SalesPointFieldExternalReferenceId] = command.ExternalReferenceId
		fields[models.SalesPointFieldExternalReferenceType] = command.ExternalReferenceType
	}

	created, err := pointEngine.ResourceService().Create(ctx, fields)
	if err != nil {
		return nil, err
	}
	if created.ClientErrors.Count() > 0 {
		return &it.CreateSalesPointResult{ClientErrors: created.ClientErrors}, nil
	}
	return &it.CreateSalesPointResult{
		HasData: true,
		Data: it.SalesPointData{
			Id:               string(*id),
			SalesChannelId:   channelId,
			SalesChannelCode: stringOf(channel, models.SalesChannelFieldCode),
		},
	}, nil
}

func (this *SalesPointApplicationServiceImpl) ArchiveSalesPoint(
	ctx corectx.Context, command it.SalesPointCommand,
) (*it.SalesPointMutationResult, error) {
	return this.mutate(ctx, command, drif.PermissionSetArchived,
		func(service *services.SalesPointDomainServiceImpl) (*dyn.OpResult[dyn.MutateResultData], error) {
			return service.Archive(ctx, command.SalesPointId)
		})
}

func (this *SalesPointApplicationServiceImpl) SuspendSalesPoint(
	ctx corectx.Context, command it.SalesPointCommand,
) (*it.SalesPointMutationResult, error) {
	return this.mutate(ctx, command, "suspend",
		func(service *services.SalesPointDomainServiceImpl) (*dyn.OpResult[dyn.MutateResultData], error) {
			return service.Suspend(ctx, command.SalesPointId)
		})
}

func (this *SalesPointApplicationServiceImpl) ActivateSalesPoint(
	ctx corectx.Context, command it.SalesPointCommand,
) (*it.SalesPointMutationResult, error) {
	return this.mutate(ctx, command, "activate",
		func(service *services.SalesPointDomainServiceImpl) (*dyn.OpResult[dyn.MutateResultData], error) {
			return service.Activate(ctx, command.SalesPointId)
		})
}

// DeleteSalesPoint removes a sales point, or archives it when it carries sales history. The caller
// is told which happened, because a second call could see a different answer.
func (this *SalesPointApplicationServiceImpl) DeleteSalesPoint(
	ctx corectx.Context, command it.SalesPointCommand,
) (*it.DeleteSalesPointResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionDelete,
		c.SalesPointResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.DeleteSalesPointResult{ClientErrors: *cErrs}, nil
	}

	service, err := salesPointService()
	if err != nil {
		return nil, err
	}
	result, err := service.DeleteOrArchive(ctx, command.SalesPointId)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.DeleteSalesPointResult{ClientErrors: result.ClientErrors}, nil
	}

	// Re-read to report the outcome: a point still present was archived, one gone was removed.
	remaining, err := loadSalesPoint(ctx, command.SalesPointId)
	if err != nil {
		return nil, err
	}
	return &it.DeleteSalesPointResult{
		HasData:  true,
		Archived: remaining != nil && boolOf(remaining, basemodel.FieldIsArchived),
	}, nil
}

func (this *SalesPointApplicationServiceImpl) mutate(
	ctx corectx.Context, command it.SalesPointCommand, permission string,
	run func(*services.SalesPointDomainServiceImpl) (*dyn.OpResult[dyn.MutateResultData], error),
) (*it.SalesPointMutationResult, error) {
	if cErrs := assertPermission(ctx, permission,
		c.SalesPointResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.SalesPointMutationResult{ClientErrors: *cErrs}, nil
	}
	service, err := salesPointService()
	if err != nil {
		return nil, err
	}
	outcome, err := run(service)
	if err != nil {
		return nil, err
	}
	if outcome.ClientErrors.Count() > 0 {
		return &it.SalesPointMutationResult{ClientErrors: outcome.ClientErrors}, nil
	}
	return &it.SalesPointMutationResult{HasData: true}, nil
}

func (this *SalesPointApplicationServiceImpl) resolveChannel(
	ctx corectx.Context, code string,
) (dmodel.DynamicFields, error) {
	normalized := services.NormalizeChannelCode(code)
	if normalized == "" {
		return nil, nil
	}
	engine, err := services.EngineFor(models.SalesChannelSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := models.FindSalesChannelByCode(ctx, engine.ResourceRepository(), normalized)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return found[0], nil
}

func salesPointService() (*services.SalesPointDomainServiceImpl, error) {
	engine, err := services.EngineFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	service, ok := engine.ResourceService().(*services.SalesPointDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales point engine is not running the derived sales point service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func loadSalesPoint(ctx corectx.Context, salesPointId string) (dmodel.DynamicFields, error) {
	engine, err := services.EngineFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.SalesPointFieldId: salesPointId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadSalesPoint")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

func pointRejection(key, message string) *it.CreateSalesPointResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesPointSchemaName, key, message))
	return &it.CreateSalesPointResult{ClientErrors: *vErrs}
}

func boolOf(record dmodel.DynamicFields, field string) bool {
	if record == nil {
		return false
	}
	value, ok := record[field]
	if !ok || value == nil {
		return false
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return false
}
