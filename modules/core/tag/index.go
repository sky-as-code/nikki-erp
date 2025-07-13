package tag

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/tag/impl"
	it "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildTagDescriptor()),
		deps.Register(func() it.TagServiceFactory {
			return tagServiceFactory
		}),
	)

	return err
}

func tagServiceFactory(tagType enum.EnumType) (tagService it.TagService, err error) {
	err = deps.Invoke(func(enumSvc enum.EnumService, eventBus event.EventBus) {
		tagService = impl.NewTagServiceImpl(enumSvc, eventBus, tagType)
	})
	return
}
