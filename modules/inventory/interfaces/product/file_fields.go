package product

import (
	"mime/multipart"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type UploadFileCommand struct {
	RecordId  model.Id
	FieldName string
	File      *multipart.FileHeader
	Params    dmodel.DynamicFields
}

type DeleteFileCommand struct {
	RecordId  model.Id
	FieldName string
	Params    dmodel.DynamicFields
}

type FileFieldResult = dyn.OpResult[dyn.MutateResultData]

type FileFieldsService interface {
	UploadFile(ctx corectx.Context, cmd UploadFileCommand) (*FileFieldResult, error)
	ReplaceFile(ctx corectx.Context, cmd UploadFileCommand) (*FileFieldResult, error)
	DeleteFile(ctx corectx.Context, cmd DeleteFileCommand) (*FileFieldResult, error)
}
