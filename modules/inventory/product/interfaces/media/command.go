package media

import (
	"mime/multipart"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type UploadMediaCommand struct {
	Source     string
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type UploadMediaResult = dynamicmodel.OpResult[domain.InventoryMedia]

type MediaData struct {
	domain.InventoryMedia
	Url string
}

func (this *MediaData) GetFieldData() dmodel.DynamicFields {
	base := this.InventoryMedia.GetFieldData()
	base["url"] = this.Url
	return base
}

type DeleteMediaCommand = dynamicmodel.DeleteOneCommand
type DeleteMediaResult = dynamicmodel.OpResult[dynamicmodel.MutateResultData]

type GetMediaQuery = dynamicmodel.GetOneQuery
type GetMediaResult = dynamicmodel.OpResult[MediaData]

type SearchMediaQuery = dynamicmodel.SearchQuery
type SearchMediaData = dynamicmodel.PagedResultData[MediaData]
type SearchMediaResult = dynamicmodel.OpResult[SearchMediaData]
