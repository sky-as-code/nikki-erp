package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// CreateRoot creates an org's root category, the parent every parent-less category is filled
// with. The root is its own parent, which is where the tree walks stop, so its id is chosen
// before the write. It skips Create's rules on purpose: those refuse is_root from a caller, fill
// the parent with a root that does not exist yet, and check that the parent exists.
func (this *ProductCategoryDomainServiceImpl) CreateRoot(
	ctx corectx.Context, cmd itProduct.CreateProductCategoryCommand,
) (*itProduct.CreateProductCategoryResult, error) {
	rootId, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "CreateRoot")
	}

	root := make(dmodel.DynamicFields, len(cmd)+4)
	for key, value := range cmd {
		root[key] = value
	}
	root[basemodel.FieldId] = string(*rootId)
	root[models.ProductCategoryFieldParentCategoryId] = string(*rootId)
	root[models.ProductCategoryFieldIsRoot] = true
	// A root's path is its name alone.
	root[models.ProductCategoryFieldCanonicalName] = cmd[models.ProductCategoryFieldName]

	root[basemodel.FieldIsArchived] = false

	// Written by the repository: Create checks that the parent exists before inserting, and a
	// row that is its own parent does not exist until it is inserted.
	inserted, err := this.Repository().Insert(ctx, root)
	if err != nil {
		return nil, errors.Wrap(err, "CreateRoot")
	}
	if inserted.ClientErrors.Count() > 0 {
		return &itProduct.CreateProductCategoryResult{ClientErrors: inserted.ClientErrors}, nil
	}
	return &itProduct.CreateProductCategoryResult{Data: root, HasData: true}, nil
}
