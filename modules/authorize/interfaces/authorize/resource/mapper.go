package resource

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateResourceCommand) ToDomainModel() *domain.Resource {
	resource := &domain.Resource{}
	model.MustCopy(this, resource)
	return resource
}

func (this UpdateResourceCommand) ToDomainModel() *domain.Resource {
	resource := &domain.Resource{}
	model.MustCopy(this, resource)
	return resource
}

func (this DeleteResourceHardByNameQuery) ToDomainModel() *domain.Resource {
	resource := &domain.Resource{}
	model.MustCopy(this, resource)
	return resource
}
