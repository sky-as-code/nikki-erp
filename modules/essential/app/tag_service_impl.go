package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

func NewTagServiceImpl(tagRepo it.TagRepository) it.TagService {
	return &TagServiceImpl{
		tagRepo: tagRepo,
	}
}

type TagServiceImpl struct {
	tagRepo it.TagRepository
}

func (this *TagServiceImpl) CreateTag(ctx corectx.Context, cmd it.CreateTagCommand) (*it.CreateTagResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Tag, *models.Tag]{
		Action:         "create tag",
		BaseRepoGetter: this.tagRepo,
		Data:           cmd,
	})
}

func (this *TagServiceImpl) UpdateTag(ctx corectx.Context, cmd it.UpdateTagCommand) (*it.UpdateTagResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Tag, *models.Tag]{
		Action:       "update tag",
		DbRepoGetter: this.tagRepo,
		Data:         cmd,
	})
}

func (this *TagServiceImpl) DeleteTag(ctx corectx.Context, cmd it.DeleteTagCommand) (*it.DeleteTagResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete tag",
		DbRepoGetter: this.tagRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *TagServiceImpl) GetTag(ctx corectx.Context, query it.GetTagQuery) (*it.GetTagResult, error) {
	return corecrud.GetOne[models.Tag](ctx, corecrud.GetOneParam{
		Action:       "get tag",
		DbRepoGetter: this.tagRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *TagServiceImpl) SearchTags(ctx corectx.Context, query it.SearchTagsQuery) (*it.SearchTagsResult, error) {
	return corecrud.Search[models.Tag](ctx, corecrud.SearchParam{
		Action:       "search tags",
		DbRepoGetter: this.tagRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *TagServiceImpl) TagExists(ctx corectx.Context, query it.TagExistsQuery) (*it.TagExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if tag exists",
		DbRepoGetter: this.tagRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
