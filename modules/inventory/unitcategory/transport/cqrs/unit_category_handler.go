package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/interfaces"
)

func NewUnitCategoryHandler(unitCategorySvc it.UnitCategoryService) *UnitCategoryHandler {
	return &UnitCategoryHandler{
		UnitCategorySvc: unitCategorySvc,
	}
}

type UnitCategoryHandler struct {
	UnitCategorySvc it.UnitCategoryService
}

func (h *UnitCategoryHandler) CreateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[it.CreateUnitCategoryCommand]) (*cqrs.Reply[it.CreateUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.CreateUnitCategory)
}

func (h *UnitCategoryHandler) UpdateUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUnitCategoryCommand]) (*cqrs.Reply[it.UpdateUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.UpdateUnitCategory)
}

func (h *UnitCategoryHandler) DeleteUnitCategory(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUnitCategoryCommand]) (*cqrs.Reply[it.DeleteUnitCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.DeleteUnitCategory)
}

func (h *UnitCategoryHandler) GetUnitCategoryById(ctx context.Context, packet *cqrs.RequestPacket[it.GetUnitCategoryByIdQuery]) (*cqrs.Reply[it.GetUnitCategoryByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.GetUnitCategoryById)
}

func (h *UnitCategoryHandler) SearchUnitCategories(ctx context.Context, packet *cqrs.RequestPacket[it.SearchUnitCategoriesQuery]) (*cqrs.Reply[it.SearchUnitCategoriesResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.UnitCategorySvc.SearchUnitCategories)
}
