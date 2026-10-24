package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/samber/lo"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateApplicationService is handed the composable default by the template onion.
func NewProductTemplateApplicationService(
	base composable.CrudApplicationService, catalogChangedPub itEvent.CatalogChangedEventPublisher, storage filestorage.FileStorageAdapter,
) itProduct.ProductTemplateApplicationService {
	productSvc, ok := base.DomainService().(itProduct.ProductService)
	if !ok {
		panic(errors.New("the product template onion must be built with NewProductTemplateDomainService"))
	}
	return &ProductTemplateApplicationServiceImpl{
		fileBackedApplicationService: newFileBackedApplicationService(
			base, storage, services.ProductTemplateFileFields()),
		productSvc:        productSvc,
		catalogChangedPub: catalogChangedPub,
	}
}

type ProductTemplateApplicationServiceImpl struct {
	fileBackedApplicationService
	productSvc        itProduct.ProductService
	catalogChangedPub itEvent.CatalogChangedEventPublisher
}

// GenerateVariants brings the template's variants in step with its attribute configuration. The
// template is fetched first: its absence is a missing record, not a malformed request.
func (this *ProductTemplateApplicationServiceImpl) GenerateVariants(
	ctx corectx.Context, cmd itProduct.GenerateTemplateVariantsCommand,
) (*itProduct.GenerateTemplateVariantsResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionGenerateVariants, cmd); err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}

	vErrs := ft.ClientErrors{}
	found, err := this.FetchByKeys(ctx, dmodel.DynamicFields{
		models.ProductTemplateFieldId: readStringField(cmd, paramRecordId),
	}, &vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return anyFailure(&vErrs, nil)
	}
	if found == nil {
		return &itProduct.GenerateTemplateVariantsResult{}, nil
	}

	templateId := derefId(models.NewProductTemplateFrom(found).GetId())
	result, err := this.productSvc.GenerateVariants(ctx, itProduct.GenerateVariantsQuery{
		TemplateId: templateId, DeleteAllowed: this.mayDeleteVariants(ctx, found),
	})
	if err != nil {
		return nil, errors.Wrap(err, "GenerateVariants")
	}
	if result.ClientErrors.Count() > 0 {
		return anyFailure(&result.ClientErrors, nil)
	}
	// Always data: "nothing to generate" is a valid answer with a payload, and no data would
	// reach the caller as a missing record.
	return anyResult(itProduct.NewGenerateVariantsView(result.Data)), nil
}

// ResolveSelection resolves a full attribute selection to its complete variant (plan D4). A
// lookup needs read; materializing a missing or configuration-archived complete writes a variant
// of the template, which is the generate_variants power.
func (this *ProductTemplateApplicationServiceImpl) ResolveSelection(
	ctx corectx.Context, query itProduct.ResolveSelectionQuery,
) (*itProduct.ResolveSelectionResult, error) {
	permission := composable.PermissionRead
	if readResolveMaterialize(query) {
		permission = PermissionGenerateVariants
	}
	orgId, cErrs := this.AssertAction(ctx, permission, query)
	if cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	resolveQuery, vErrs := buildResolveCompleteQuery(query)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}
	if orgId != nil {
		inOrg := dmodel.DynamicFields{models.ProductTemplateFieldId: resolveQuery.TemplateId}
		if cErrs, err := this.AssertRecordInOrg(ctx, inOrg, *orgId); err != nil || cErrs != nil {
			return anyFailure(cErrs, err)
		}
	}

	resolver, ok := this.productSvc.(itProduct.CompleteVariantResolver)
	if !ok {
		return nil, errors.New("ResolveSelection: the product service cannot resolve complete variants")
	}
	result, err := resolver.ResolveCompleteVariant(ctx, resolveQuery)
	if err != nil {
		return nil, errors.Wrap(err, "ResolveSelection")
	}
	if result.ClientErrors.Count() > 0 {
		return anyFailure(&result.ClientErrors, nil)
	}
	// No data means the template does not exist, which the REST layer turns into a missing record.
	if !result.HasData {
		return &itProduct.ResolveSelectionResult{}, nil
	}
	return anyResult(itProduct.NewResolveCompleteVariantView(result.Data)), nil
}

func (this *ProductTemplateApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	this.stampVariantDeleteAllowed(ctx, cmd)
	res, err := this.fileBackedApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	this.stampVariantDeleteAllowed(ctx, cmd)
	res, err := this.fileBackedApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) UploadFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.UploadFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) ReplaceFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.ReplaceFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}

func (this *ProductTemplateApplicationServiceImpl) DeleteFile(
	ctx corectx.Context, cmd itProduct.DeleteFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.DeleteFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductTemplate, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}
