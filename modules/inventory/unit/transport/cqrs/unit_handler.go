package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
)

func NewUnitHandler(unitSvc it.UnitService) *UnitHandler {
	return &UnitHandler{
		UnitSvc: unitSvc,
	}
}

type UnitHandler struct {
	UnitSvc it.UnitService
}

func (this *UnitHandler) CreateUnit(ctx context.Context, packet *cqrs.RequestPacket[it.CreateUnitCommand]) (*cqrs.Reply[it.CreateUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.CreateUnit)
}

func (this *UnitHandler) UpdateUnit(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUnitCommand]) (*cqrs.Reply[it.UpdateUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.UpdateUnit)
}

func (this *UnitHandler) DeleteUnit(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUnitCommand]) (*cqrs.Reply[it.DeleteUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.DeleteUnit)
}

func (this *UnitHandler) GetUnitById(ctx context.Context, packet *cqrs.RequestPacket[it.GetUnitByIdQuery]) (*cqrs.Reply[it.GetUnitByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.GetUnitById)
}

func (this *UnitHandler) SearchUnits(ctx context.Context, packet *cqrs.RequestPacket[it.SearchUnitsQuery]) (*cqrs.Reply[it.SearchUnitsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.SearchUnits)
}
