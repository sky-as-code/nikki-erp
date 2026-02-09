package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

func NewUnitCategoryHandler(unitCategorySvc itUnitCategory.UnitCategoryService) *UnitCategoryHandler {
	return &UnitCategoryHandler{
		UnitCategorySvc: unitCategorySvc,
	}
}

type UnitCategoryHandler struct {
	UnitCategorySvc itUnitCategory.UnitCategoryService
}

func (h *UnitCategoryHandler) CreateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.CreateUnitCategoryCommand]) (*cqrs.Reply[itUnitCategory.CreateUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.CreateUnitCategory)
}

func (h *UnitCategoryHandler) UpdateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.UpdateUnitCategoryCommand]) (*cqrs.Reply[itUnitCategory.UpdateUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.UpdateUnitCategory)
}

func (h *UnitCategoryHandler) DeleteUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.DeleteUnitCategoryCommand]) (*cqrs.Reply[itUnitCategory.DeleteUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.DeleteUnitCategory)
}

func (h *UnitCategoryHandler) GetUnitCategoryById(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.GetUnitCategoryByIdQuery]) (*cqrs.Reply[itUnitCategory.GetUnitCategoryByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.GetUnitCategoryById)
}

func (h *UnitCategoryHandler) SearchUnitCategories(ctx context.Context, packet *cqrs.RequestPacket[itUnitCategory.SearchUnitCategoriesQuery]) (*cqrs.Reply[itUnitCategory.SearchUnitCategoriesResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.SearchUnitCategories)
}
