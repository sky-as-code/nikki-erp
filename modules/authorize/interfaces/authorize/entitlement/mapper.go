package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateEntitlementCommand) ToEntitlement() *domain.Entitlement {
	entitlement := &domain.Entitlement{}
	model.MustCopy(this, entitlement)

	return entitlement
}

func (this UpdateEntitlementCommand) ToEntitlement() *domain.Entitlement {
	entitlement := &domain.Entitlement{}
	model.MustCopy(this, entitlement)

	return entitlement
}
