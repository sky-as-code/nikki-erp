package entitlement

// import (
// 	"github.com/sky-as-code/nikki-erp/common/model"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// )

// func (this CreateEntitlementCommand) ToDomainModel() *domain.Entitlement {
// 	e := domain.NewEntitlement()
// 	e.SetName(&this.Name)
// 	e.SetDescription(this.Description)
// 	expr := this.ActionExpr
// 	e.SetActionExpr(&expr)
// 	cb := model.Id(this.CreatedBy)
// 	e.SetCreatedBy(&cb)
// 	if this.ActionId != nil {
// 		aid := model.Id(*this.ActionId)
// 		e.SetActionId(&aid)
// 	}
// 	if this.ResourceId != nil {
// 		rid := model.Id(*this.ResourceId)
// 		e.SetResourceId(&rid)
// 	}
// 	return e
// }

// func (this UpdateEntitlementCommand) ToDomainModel() *domain.Entitlement {
// 	e := domain.NewEntitlement()
// 	id := model.Id(this.Id)
// 	e.SetId(&id)
// 	e.SetEtag(this.Etag)
// 	e.SetDescription(this.Description)
// 	return e
// }

// func (this DeleteEntitlementHardByIdCommand) ToDomainModel() *domain.Entitlement {
// 	e := domain.NewEntitlement()
// 	id := model.Id(this.Id)
// 	e.SetId(&id)
// 	return e
// }
