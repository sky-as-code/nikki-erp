package entitlement_assignment

// import (
// 	"github.com/sky-as-code/nikki-erp/common/model"

// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// )

// func (this CreateEntitlementAssignmentCommand) ToDomainModel() *domain.EntitlementGrant {
// 	a := domain.NewEntitlementAssignment()
// 	a.SetEntitlementId(&this.EntitlementId)
// 	st := this.SubjectType
// 	a.SetSubjectType(&st)
// 	sref := string(this.SubjectRef)
// 	a.SetSubjectRef(&sref)
// 	a.SetActionName(this.ActionName)
// 	a.SetResourceName(this.ResourceName)
// 	re := this.ResolvedExpr
// 	a.SetResolvedExpr(&re)
// 	a.SetScopeRef(this.ScopeRef)
// 	return a
// }

// func (this DeleteEntitlementAssignmentByIdCommand) ToDomainModel() *domain.EntitlementGrant {
// 	a := domain.NewEntitlementAssignment()
// 	id := model.Id(this.Id)
// 	a.SetId(&id)
// 	return a
// }

// func (this DeleteEntitlementAssignmentByEntitlementIdCommand) ToDomainModel() *domain.EntitlementGrant {
// 	a := domain.NewEntitlementAssignment()
// 	a.SetEntitlementId(&this.EntitlementId)
// 	return a
// }
