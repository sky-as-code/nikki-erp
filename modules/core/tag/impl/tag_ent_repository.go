package impl

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/enum/impl"
)

func BuildTagDescriptor() *orm.EntityDescriptor {
	return impl.GetEnumDescriptorBuilder("tag").
		Aliases("tags").
		Descriptor()
}
