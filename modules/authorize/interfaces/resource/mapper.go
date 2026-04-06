package resource

// import (
// 	"github.com/sky-as-code/nikki-erp/common/model"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// )

// func (this CreateResourceCommand) ToDomainModel() *domain.Resource {
// 	r := domain.NewResource()
// 	r.SetName(&this.Name)
// 	r.SetDescription(this.Description)
// 	rt := domain.ResourceOwnerType(this.ResourceType)
// 	r.SetResourceType(&rt)
// 	ref := this.ResourceRef
// 	r.SetResourceRef(&ref)
// 	st := domain.ResourceScope(this.ScopeType)
// 	r.SetScopeType(&st)
// 	return r
// }

// func (this UpdateResourceCommand) ToDomainModel() *domain.Resource {
// 	r := domain.NewResource()
// 	id := model.Id(this.Id)
// 	r.SetId(&id)
// 	r.SetEtag(this.Etag)
// 	r.SetDescription(this.Description)
// 	return r
// }

// func (this DeleteResourceHardByNameQuery) ToDomainModel() *domain.Resource {
// 	r := domain.NewResource()
// 	n := this.Name
// 	r.SetName(&n)
// 	return r
// }
