package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

func NewUnitHandler(unitSvc itUnit.UnitService, logger logging.LoggerService) *UnitHandler {
	return &UnitHandler{
		Logger:  logger,
		UnitSvc: unitSvc,
	}
}

type UnitHandler struct {
	Logger  logging.LoggerService
	UnitSvc itUnit.UnitService
}

func (this *UnitHandler) CreateUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.CreateUnitCommand]) (
	*cqrs.Reply[itUnit.CreateUnitResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitSvc.CreateUnit)
}

func (this *UnitHandler) UpdateUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.UpdateUnitCommand]) (
	*cqrs.Reply[itUnit.UpdateUnitResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitSvc.UpdateUnit)
}

func (this *UnitHandler) DeleteUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.DeleteUnitCommand]) (
	*cqrs.Reply[itUnit.DeleteUnitResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitSvc.DeleteUnit)
}

func (this *UnitHandler) GetUnit(ctx context.Context, packet *cqrs.RequestPacket[itUnit.GetUnitQuery]) (
	*cqrs.Reply[itUnit.GetUnitResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitSvc.GetUnit)
}

func (this *UnitHandler) SearchUnits(ctx context.Context, packet *cqrs.RequestPacket[itUnit.SearchUnitsQuery]) (
	*cqrs.Reply[itUnit.SearchUnitsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.UnitSvc.SearchUnits)
}
