package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	it "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

func NewTagServiceImpl(
	enumSvc enum.EnumService,
	eventBus event.EventBus,
	tagType string,
) it.TagService {
	return &TagServiceImpl{
		enumSvc:  enumSvc,
		eventBus: eventBus,
		tagType:  tagType,
	}
}

type TagServiceImpl struct {
	enumSvc  enum.EnumService
	eventBus event.EventBus
	tagType  string
}

func (this *TagServiceImpl) CreateTag(ctx crud.Context, cmd it.CreateTagCommand) (result *it.CreateTagResult, err error) {
	enumCmd := cmd.ToEnumCommand(this.tagType)
	tag, err := this.enumSvc.CreateEnum(ctx, enumCmd)
	ft.PanicOnErr(err)

	return it.NewCreateTagResult(tag), err
}

func (this *TagServiceImpl) UpdateTag(ctx crud.Context, cmd it.UpdateTagCommand) (result *it.UpdateTagResult, err error) {
	enumCmd := cmd.ToEnumCommand()
	tag, err := this.enumSvc.UpdateEnum(ctx, enumCmd)
	return it.NewUpdateTagResult(tag), err
}

func (this *TagServiceImpl) DeleteTag(ctx crud.Context, cmd it.DeleteTagCommand) (result *it.DeleteTagResult, err error) {
	enumCmd := cmd.ToEnumCommand(this.tagType)
	tag, err := this.enumSvc.DeleteEnum(ctx, enumCmd)
	return it.NewDeleteTagResult(tag), err
}

func (this *TagServiceImpl) TagExistsMulti(ctx crud.Context, cmd it.TagExistsMultiQuery) (result *it.TagExistsMultiResult, err error) {
	enumCmd := cmd.ToEnumQuery()
	enumResult, err := this.enumSvc.EnumExistsMulti(ctx, enumCmd)
	return it.NewTagExistsMultiResult(enumResult), err
}

func (this *TagServiceImpl) GetTagById(ctx crud.Context, query it.GetTagByIdQuery) (result *it.GetTagByIdResult, err error) {
	enumCmd := query.ToEnumQuery()
	tag, err := this.enumSvc.GetEnum(ctx, enumCmd)
	return it.NewGetTagByIdResult(tag), err
}

func (this *TagServiceImpl) ListTags(ctx crud.Context, query it.ListTagsQuery) (result *it.ListTagsResult, err error) {
	enumCmd := query.ToEnumQuery(this.tagType)
	tags, err := this.enumSvc.ListEnums(ctx, enumCmd)
	return it.NewListTagsResult(tags), err
}

func (this *TagServiceImpl) SearchTags(ctx crud.Context, query it.SearchTagsQuery) (result *it.SearchTagsResult, err error) {
	enumCmd := query.ToEnumQuery(this.tagType)
	tags, err := this.enumSvc.SearchEnums(ctx, enumCmd)
	return it.NewListTagsResult(tags), err
}
