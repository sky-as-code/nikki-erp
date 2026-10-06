package product

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type FileFieldsService interface {
	ClearFiles(
		ctx corectx.Context, recordId model.Id, params dmodel.DynamicFields,
	) (*dyn.OpResult[dyn.MutateResultData], error)
}
