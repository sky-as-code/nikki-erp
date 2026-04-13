package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

func NewUnitCategoryHandler(unitCategorySvc itUnitCategory.UnitCategoryService, logger logging.LoggerService) *UnitCategoryHandler {
	return &UnitCategoryHandler{
		Logger:          logger,
		UnitCategorySvc: unitCategorySvc,
	}
}

type UnitCategoryHandler struct {
	Logger          logging.LoggerService
	UnitCategorySvc itUnitCategory.UnitCategoryService
}

func (this *UnitCategoryHandler) CreateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.CreateUnitCategoryCommand]) (
	*cqrs.Reply[itUnitCategory.CreateUnitCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitCategorySvc.CreateUnitCategory)
}

func (this *UnitCategoryHandler) UpdateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.UpdateUnitCategoryCommand]) (
	*cqrs.Reply[itUnitCategory.UpdateUnitCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitCategorySvc.UpdateUnitCategory)
}

func (this *UnitCategoryHandler) DeleteUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.DeleteUnitCategoryCommand]) (
	*cqrs.Reply[itUnitCategory.DeleteUnitCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitCategorySvc.DeleteUnitCategory)
}

func (this *UnitCategoryHandler) GetUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.GetUnitCategoryQuery]) (
	*cqrs.Reply[itUnitCategory.GetUnitCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitCategorySvc.GetUnitCategory)
}

func (this *UnitCategoryHandler) SearchUnitCategories(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.SearchUnitCategoriesQuery]) (
	*cqrs.Reply[itUnitCategory.SearchUnitCategoriesResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitCategorySvc.SearchUnitCategories)
}
