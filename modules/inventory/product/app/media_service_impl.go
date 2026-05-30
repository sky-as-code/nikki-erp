package app

import (
	"time"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/objectkey"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/upload"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/media"
)

func NewInventoryMediaServiceImpl(
	repo it.InventoryMediaRepository,
	storage filestorage.FileStorageAdapter,
) it.InventoryMediaService {
	return &InventoryMediaServiceImpl{repo, storage}
}

type InventoryMediaServiceImpl struct {
	repo    it.InventoryMediaRepository
	storage filestorage.FileStorageAdapter
}

func (this *InventoryMediaServiceImpl) Upload(ctx corectx.Context, cmd it.UploadMediaCommand) (*it.UploadMediaResult, error) {
	var err error
	if cmd.File == nil {
		return nil, nil
	}

	entity := domain.NewInventoryMedia()

	storageKey, err := objectkey.BuildFromFileHeader(cmd.Source, cmd.FileHeader)
	if err != nil {
		return nil, err
	}

	mediaType, uploadReader, err := upload.SniffContentTypeAndRewind(cmd.File, cmd.FileHeader)
	if err != nil {
		return nil, err
	}

	entity.SetStorageKey(&storageKey)
	entity.SetMediaType(&mediaType)
	entity.SetResource(&cmd.Source)

	size := int64(0)
	if cmd.FileHeader != nil {
		size = cmd.FileHeader.Size
	}
	if err = this.storage.Put(ctx.InnerContext(),
		storageKey,
		uploadReader,
		filestorage.NewPutOptions(mediaType, size)); err != nil {
		return nil, err
	}

	return corecrud.Create(ctx, corecrud.CreateParam[domain.InventoryMedia, *domain.InventoryMedia]{
		Action:         "create media",
		Data:           entity,
		BaseRepoGetter: this.repo,
	})
}

func (this *InventoryMediaServiceImpl) Delete(ctx corectx.Context, cmd it.DeleteMediaCommand) (*it.DeleteMediaResult, error) {
	res := &it.DeleteMediaResult{}
	getMediaRes, err := this.GetMedia(ctx, it.GetMediaQuery{
		Id: cmd.Id,
	})
	if err != nil {
		return nil, nil
	}

	if getMediaRes.ClientErrors.Count() > 0 {
		res.ClientErrors = getMediaRes.ClientErrors
		return res, nil
	}

	if !getMediaRes.HasData {
		res.ClientErrors = fault.ClientErrors{*fault.NewNotFoundError("id")}
		return res, nil
	}

	data := getMediaRes.Data
	key := lo.FromPtr(data.GetStorageKey())

	err = this.storage.Remove(ctx, key)
	if err != nil {
		pendingDelete := true
		data.SetPendingDelete(&pendingDelete)
		return crud.Update(ctx, crud.UpdateParam[domain.InventoryMedia, *domain.InventoryMedia]{
			Action:       "Update Media",
			DbRepoGetter: this.repo,
			Data:         data.InventoryMedia,
		})
	}

	return crud.DeleteOne(ctx, crud.DeleteOneParam{
		Action:       "Delete Media",
		DbRepoGetter: this.repo,
		Cmd:          cmd,
	})
}

func (this *InventoryMediaServiceImpl) GetMedia(ctx corectx.Context, query it.GetMediaQuery) (*it.GetMediaResult, error) {
	res := &it.GetMediaResult{
		Data:    it.MediaData{},
		HasData: false,
	}

	getOneRes, err := corecrud.GetOne[domain.InventoryMedia](ctx, corecrud.GetOneParam{
		Action:       "Get Media",
		DbRepoGetter: this.repo,
		Query:        query,
	})

	if err != nil {
		return nil, err
	}

	if getOneRes.ClientErrors.Count() > 0 {
		res.ClientErrors = getOneRes.ClientErrors
		return res, nil
	}

	if !getOneRes.HasData {
		res.ClientErrors = fault.ClientErrors{*fault.NewNotFoundError("media_id")}
		return res, nil
	}

	data := getOneRes.Data

	res.Data.InventoryMedia = data
	res.HasData = true

	key := data.GetStorageKey()
	if key == nil {
		return res, nil
	}

	url, err := this.storage.GeneratePresignedURL(ctx, *key, time.Hour)
	if err != nil {
		return res, err
	}

	res.Data.Url = url

	return res, nil
}

func (this *InventoryMediaServiceImpl) SearchMedia(ctx corectx.Context, query it.SearchMediaQuery) (*it.SearchMediaResult, error) {
	res := &it.SearchMediaResult{}
	searchRes, err := crud.Search[domain.InventoryMedia](ctx, corecrud.SearchParam{
		Action:       "Search Media",
		DbRepoGetter: this.repo,
		Query:        query,
	})
	if err != nil {
		return nil, err
	}

	if searchRes.ClientErrors.Count() > 0 {
		res.ClientErrors = searchRes.ClientErrors
		return res, nil
	}

	if !searchRes.HasData {
		return res, nil
	}

	items := searchRes.Data.Items
	keys := make([]string, 0, len(items))
	for _, item := range items {
		key := item.GetStorageKey()
		if key == nil {
			continue
		}

		keys = append(keys, *key)
	}

	urls, err := filestorage.GeneratePresignedBulk(ctx, this.storage, keys, time.Hour)
	if err != nil {
		return res, err
	}

	resItems := make([]it.MediaData, 0, len(items))
	for _, item := range items {
		resItem := it.MediaData{
			InventoryMedia: item,
		}

		if key := item.GetStorageKey(); key != nil {
			if url, ok := urls[*key]; ok {
				resItem.Url = url
			}
		}

		resItems = append(resItems, resItem)
	}

	res.HasData = true
	res.Data.Items = resItems
	res.Data.Page = searchRes.Data.Page
	res.Data.Size = searchRes.Data.Size
	res.Data.Total = searchRes.Data.Total

	return res, nil
}

