package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	it "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
)

type SalesChannelApplicationServiceImpl struct{}

func NewSalesChannelApplicationServiceImpl() it.SalesChannelAppService {
	return &SalesChannelApplicationServiceImpl{}
}

// RegisterSalesChannel lets another module claim a channel. It is idempotent by code, since modules
// run it on every boot, and refuses a code already owned by a different module rather than handing
// over somebody else's integration.
func (this *SalesChannelApplicationServiceImpl) RegisterSalesChannel(
	ctx corectx.Context, command it.RegisterSalesChannelCommand,
) (*it.RegisterSalesChannelResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionCreate,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.RegisterSalesChannelResult{ClientErrors: *cErrs}, nil
	}

	code := services.NormalizeChannelCode(command.Code)
	if !services.IsValidChannelCode(code) {
		return channelRejection("sales_channel.code_invalid",
			"a sales channel code must be lowercase alphanumeric with no separators"), nil
	}
	if command.ManagedByModule == "" {
		return channelRejection("sales_channel.managed_by_module_required",
			"registering a sales channel requires the name of the module that owns it"), nil
	}

	engine, err := services.EngineFor(models.SalesChannelSchemaName)
	if err != nil {
		return nil, err
	}

	existing, err := models.FindSalesChannelByCode(ctx, engine.ResourceRepository(), code)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		owner := stringOf(existing[0], models.SalesChannelFieldManagedByModule)
		if owner != command.ManagedByModule {
			return channelRejection("sales_channel.code_already_owned",
				"the sales channel code '"+code+"' is registered by module '"+owner+"'"), nil
		}
		// Same owner, same code: the retry path must look exactly like success.
		return &it.RegisterSalesChannelResult{
			HasData: true,
			Data: it.SalesChannelData{
				Id:   stringOf(existing[0], models.SalesChannelFieldId),
				Code: code,
			},
		}, nil
	}

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "RegisterSalesChannel")
	}
	created, err := engine.ResourceService().Create(ctx, dmodel.DynamicFields{
		models.SalesChannelFieldId:              string(*id),
		models.SalesChannelFieldCode:            code,
		models.SalesChannelFieldName:            command.Name,
		models.SalesChannelFieldDescription:     command.Description,
		models.SalesChannelFieldManagedByModule: command.ManagedByModule,
		models.SalesChannelFieldStatus:          string(models.SalesChannelStatusActive),
		models.SalesChannelFieldIsSystem:        false,
	})
	if err != nil {
		return nil, err
	}
	if created.ClientErrors.Count() > 0 {
		return &it.RegisterSalesChannelResult{ClientErrors: created.ClientErrors}, nil
	}
	return &it.RegisterSalesChannelResult{
		HasData: true,
		Data:    it.SalesChannelData{Id: string(*id), Code: code},
	}, nil
}

func (this *SalesChannelApplicationServiceImpl) ResolveSalesChannelByCode(
	ctx corectx.Context, query it.ResolveSalesChannelQuery,
) (*it.ResolveSalesChannelResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ResolveSalesChannelResult{ClientErrors: *cErrs}, nil
	}

	code := services.NormalizeChannelCode(query.Code)
	if code == "" {
		return &it.ResolveSalesChannelResult{
			ClientErrors: *resolveRejection("sales_channel.code_required",
				"resolving a sales channel requires a code"),
		}, nil
	}

	engine, err := services.EngineFor(models.SalesChannelSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := models.FindSalesChannelByCode(ctx, engine.ResourceRepository(), code)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return &it.ResolveSalesChannelResult{
			ClientErrors: *resolveRejection("sales_channel.not_found",
				"no sales channel with code '"+code+"'"),
		}, nil
	}

	channel := models.NewSalesChannelFrom(found[0])
	return &it.ResolveSalesChannelResult{
		HasData: true,
		Data: it.ResolvedSalesChannel{
			Id:              stringOf(found[0], models.SalesChannelFieldId),
			Code:            code,
			Name:            stringOf(found[0], models.SalesChannelFieldName),
			ManagedByModule: stringOf(found[0], models.SalesChannelFieldManagedByModule),
			Status:          stringOf(found[0], models.SalesChannelFieldStatus),
			IsUsable:        channel.IsActive(),
		},
	}, nil
}

func channelRejection(key, message string) *it.RegisterSalesChannelResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesChannelSchemaName, key, message))
	return &it.RegisterSalesChannelResult{ClientErrors: *vErrs}
}

func resolveRejection(key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesChannelSchemaName, key, message))
	return vErrs
}

func stringOf(record dmodel.DynamicFields, field string) string {
	if record == nil {
		return ""
	}
	value, ok := record[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}
