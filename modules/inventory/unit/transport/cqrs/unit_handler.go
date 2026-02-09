package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

func NewUnitHandler(unitSvc itUnit.UnitService) *UnitHandler {
	return &UnitHandler{
		UnitSvc: unitSvc,
	}
}

type UnitHandler struct {
	UnitSvc itUnit.UnitService
}

func (this *UnitHandler) CreateUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.CreateUnitCommand]) (*cqrs.Reply[itUnit.CreateUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.CreateUnit)
}

func (this *UnitHandler) UpdateUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.UpdateUnitCommand]) (*cqrs.Reply[itUnit.UpdateUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.UpdateUnit)
}

func (this *UnitHandler) DeleteUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.DeleteUnitCommand]) (*cqrs.Reply[itUnit.DeleteUnitResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.DeleteUnit)
}

func (this *UnitHandler) GetUnitById(ctx context.Context, packet *cqrs.RequestPacket[itUnit.GetUnitByIdQuery]) (*cqrs.Reply[itUnit.GetUnitByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.GetUnitById)
}

func (this *UnitHandler) SearchUnits(ctx context.Context, packet *cqrs.RequestPacket[itUnit.SearchUnitsQuery]) (*cqrs.Reply[itUnit.SearchUnitsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UnitSvc.SearchUnits)
}
