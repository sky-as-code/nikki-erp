package app

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewEntitlementServiceImpl(entitlementRepo it.EntitlementRepository, eventBus event.EventBus) it.EntitlementService {
	return &EntitlementServiceImpl{
		entitlementRepo: entitlementRepo,
		eventBus:        eventBus,
	}
}

type EntitlementServiceImpl struct {
	entitlementRepo it.EntitlementRepository
	eventBus        event.EventBus
}

func (this *EntitlementServiceImpl) CreateEntitlement(ctx context.Context, cmd it.CreateEntitlementCommand) (result *it.CreateEntitlementResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create entitlement"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()
	err = entitlement.SetDefaults()
	ft.PanicOnErr(err)
	entitlement.SetCreatedAt(time.Now())

	vErrs := entitlement.Validate(false)
	this.assertEntitlementUnique(ctx, entitlement, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement, err = this.entitlementRepo.Create(ctx, *entitlement)
	ft.PanicOnErr(err)

	return &it.CreateEntitlementResult{Data: entitlement}, err
}

func (this *EntitlementServiceImpl) assertEntitlementUnique(ctx context.Context, entitlement *domain.Entitlement, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}
	dbEntitlement, err := this.entitlementRepo.FindByName(ctx, it.FindByNameParam{Name: *entitlement.Name})
	ft.PanicOnErr(err)

	if dbEntitlement != nil {
		errors.Append("name", "name already exists")
	}
}
