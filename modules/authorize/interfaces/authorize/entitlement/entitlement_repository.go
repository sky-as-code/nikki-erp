package entitlement

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type EntitlementRepository interface {
	Create(ctx context.Context, entitlement domain.Entitlement) (*domain.Entitlement, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Entitlement, error)
}

type FindByNameParam = GetEntitlementByNameCommand